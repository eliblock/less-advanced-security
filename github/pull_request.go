package github

import (
	"bufio"
	"context"
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
				if annotation.startLine >= bound.start && annotation.startLine <= bound.end ||
					annotation.endLine >= bound.start && annotation.endLine <= bound.end {
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
		// patch header lines are formatted like:
		// @@ -0,0 +1,5 @@ <arbitrary line of code which may be blank>
		if len(line) >= 15 && strings.HasPrefix(line, "@@ -") && strings.Contains(line[2:], " @@") {
			// split into four pieces: (0) @@, (1) old line number and offset, (2), new line number and offset, (3) @@
			segments := strings.Split(line, " ")
			if len(segments) < 4 {
				continue
			}

			// split and parse new line numbers
			bounds := strings.Split(segments[2], ",")
			start, err := strconv.Atoi(bounds[0][1:]) // drop the "+"
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert %v to integer while processing %q", bounds[0][1:], line)
			}
			endOffset, err := strconv.Atoi(bounds[1])
			if err != nil {
				return nil, errors.Wrapf(err, "failed to convert %v to integer while processing %q", bounds[0][1:], line)
			}

			lineBounds = append(lineBounds, lineBound{start: start, end: start + endOffset})
		}
	}

	return lineBounds, nil
}
