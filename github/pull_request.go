package github

import (
	"context"

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

type pullRequestFile struct {
	filename string
	patch    string
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
	pr.files = append(pr.files, sdkFilesToInternalFiles(files)...)

	for response.NextPage != 0 {
		desiredPage := response.NextPage
		files, response, err = client.PullRequests.ListFiles(context.Background(), pr.owner, pr.repo, pr.number, &github.ListOptions{Page: desiredPage})
		if err != nil {
			return errors.Wrapf(err, "failed to list files (page %d)", desiredPage)
		}
		pr.files = append(pr.files, sdkFilesToInternalFiles(files)...)
	}

	return nil
}

/* * * * * Helpers * * * * */

func sdkFilesToInternalFiles(sdkFiles []*github.CommitFile) (internalFiles []*pullRequestFile) {
	for _, file := range sdkFiles {
		if file == nil || file.Filename == nil || file.Patch == nil {
			continue
		}
		internalFiles = append(internalFiles, &pullRequestFile{filename: *file.Filename, patch: *file.Patch})
	}
	return
}
