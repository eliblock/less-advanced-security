package github

import (
	"bufio"
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

type PullRequestConfiguration struct {
	Owner, Repo string
	Number      int
}

type pullRequest struct {
	owner, repo string
	number      int
	details     *github.PullRequest
	headSHA     string
	files       []*pullRequestFile
}

type lineBound struct {
	start, end int
}
type pullRequestFile struct {
	filename, patch string
	lineBounds      []lineBound
}

func createPullRequest(client *github.Client, owner string, repo string, number int, headSHA string) (*pullRequest, error) {
	pr := &pullRequest{
		owner:   owner,
		repo:    repo,
		number:  number,
		details: nil,
		headSHA: headSHA,
	}
	if err := pr.loadFromGitHub(client); err != nil {
		return nil, errors.Wrap(err, "failed to load pull request from GitHub")
	}

	// Load all files from PR (uses _latest_ files on PR, not limited by headSHA)
	if err := pr.loadFiles(client); err != nil {
		return nil, errors.Wrap(err, "failed to load files from GitHub")
	}

	return pr, nil
}

func (pr *pullRequest) loadFromGitHub(client *github.Client) error {
	details, _, err := client.PullRequests.Get(context.Background(), pr.owner, pr.repo, pr.number)
	if err != nil {
		return err
	}
	pr.details = details
	return nil
}

func (pr *pullRequest) loadFiles(client *github.Client) error {
	files, response, err := client.PullRequests.ListFiles(context.Background(), pr.owner, pr.repo, pr.number, nil)
	if err != nil {
		return errors.Wrap(err, "failed to list files (page 0)")
	}
	internalFiles, err := sdkFilesToInternalFiles(files)
	if err != nil {
		return errors.Wrap(err, "failed to convert files (page 0)")
	}
	pr.files = append(pr.files, internalFiles...)

	for response.NextPage != 0 {
		desiredPage := response.NextPage
		files, response, err = client.PullRequests.ListFiles(context.Background(), pr.owner, pr.repo, pr.number, &github.ListOptions{Page: desiredPage})
		if err != nil {
			return errors.Wrapf(err, "failed to list files (page %d)", desiredPage)
		}

		internalFiles, err := sdkFilesToInternalFiles(files)
		if err != nil {
			return errors.Wrapf(err, "failed to convert files (page %d)", desiredPage)
		}
		pr.files = append(pr.files, internalFiles...)
	}

	return nil
}

func (pr *pullRequest) filterAnnotations(annotations []*Annotation) []*Annotation {
	fileToLineBounds := make(map[string][]lineBound)
	for _, file := range pr.files {
		fileToLineBounds[file.filename] = file.lineBounds
	}

	var filteredAnnotations []*Annotation
	for _, annotation := range annotations {
		lineBounds, found := fileToLineBounds[annotation.fileName]
		if found {
			for _, bound := range lineBounds {
				start_in_bounds := annotation.startLine >= bound.start && annotation.startLine <= bound.end
				end_in_bounds := annotation.endLine >= bound.start && annotation.endLine <= bound.end
				covering_bounds := annotation.startLine <= bound.end && annotation.endLine >= bound.start
				if start_in_bounds || end_in_bounds || covering_bounds {
					filteredAnnotations = append(filteredAnnotations, annotation)
					break
				}
			}
		}
	}

	return filteredAnnotations
}

/* * * * * Helpers * * * * */

func sdkFilesToInternalFiles(sdkFiles []*github.CommitFile) ([]*pullRequestFile, error) {
	var internalFiles []*pullRequestFile

	for _, file := range sdkFiles {
		if file == nil || file.Filename == nil || file.Patch == nil {
			continue
		}

		lineBounds, err := patchToLineBounds(*file.Patch)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to generate line bounds for file %q", *file.Filename)
		}

		internalFiles = append(
			internalFiles,
			&pullRequestFile{
				filename:   *file.Filename,
				patch:      *file.Patch,
				lineBounds: lineBounds,
			},
		)
	}
	return internalFiles, nil
}

var patchLinesRegexp = regexp.MustCompile(
	`^@@ -(?P<old_start>\d+),(?P<old_rows>\d+) \+(?P<new_start>\d+),(?P<new_rows>\d) @@.*$`,
)

// Convert a file patch to an array of lineBounds.
//
// A couple optimization are used specific to our use case:
//   - only the _new_ line numbers are recorded (we will never annotate deleted
//     lines - only new or unchanged ones)
//   - the full patch range is recorded, including leading/trailing lines that
//     are unchanged (we may annotate nearby lines that were not modified)
func patchToLineBounds(patch string) ([]lineBound, error) {
	var lineBounds []lineBound

	scanner := bufio.NewScanner(strings.NewReader(patch))
	for scanner.Scan() {
		line := scanner.Text()
		match := namedMatch(patchLinesRegexp, line)
		if match == nil {
			continue
		}

		start, err := strconv.Atoi(match["new_start"])
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert %v to integer while processing %q", match["new_start"], line)
		}
		endOffset, err := strconv.Atoi(match["new_rows"]) // one-indexed offset (subtract 1 when using)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to convert %v to integer while processing %q", match["new_rows"], line)
		}

		lineBounds = append(lineBounds, lineBound{
			start: start,
			end: start + endOffset - 1,
		})
	}

	return lineBounds, nil
}

func namedMatch(re *regexp.Regexp, text string) map[string]string {
	matches := patchLinesRegexp.FindStringSubmatch(text)
	if matches == nil {
		return nil
	}
	matchesByName := map[string]string{}
	for i, name := range re.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		matchesByName[name] = matches[i]
	}
	return matchesByName
}
