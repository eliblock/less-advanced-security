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
		return nil, errors.Wrap(err, "failed to load from GitHub")
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
