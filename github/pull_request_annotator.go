package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

type PullRequestAnnotator struct {
	client *github.Client
	pr     *pullRequest
}

func CreatePullRequestAnnotator(configuration ClientConfiguration, pullRequestConfiguration PullRequestConfiguration, headSHA string) (*PullRequestAnnotator, error) {
	client, err := createClient(configuration)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create client")
	}

	pr, err := createPullRequest(client, pullRequestConfiguration.Owner, pullRequestConfiguration.Repo, pullRequestConfiguration.Number, headSHA)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pull request")
	}
	return &PullRequestAnnotator{client: client, pr: pr}, nil
}

func computeConclusion(annotations []*Annotation) string {
	// Can be one of "success", "failure", "neutral", "cancelled", "skipped", "timed_out", or "action_required".
	conclusion := "success"

	for _, annotation := range annotations {
		switch annotation.level {
		case failureLevel:
			return "failure"
		case warningLevel:
			conclusion = "neutral"
		}
	}
	return conclusion
}

func (annotator *PullRequestAnnotator) PostAnnotations(annotations []*Annotation, checkName string, filterAnnotations bool) error {
	if filterAnnotations {
		annotations = annotator.pr.filterAnnotations(annotations)
	}

	const MAX_ANNOTATIONS_PER_PAGE = 50
	// When creating a check run you can add only 50 annotations - later annotations must be added via an update to the
	// run. Split our annotations accordingly, and pull the github annotation off them.
	chunkedGitHubAnnotations := [][]*github.CheckRunAnnotation{}
	for i := 0; i < len(annotations); i += MAX_ANNOTATIONS_PER_PAGE {
		chunkEnd := i + MAX_ANNOTATIONS_PER_PAGE
		if chunkEnd > len(annotations) {
			chunkEnd = len(annotations)
		}

		thisChunk := []*github.CheckRunAnnotation{}
		for j := i; j < chunkEnd; j++ {
			thisChunk = append(thisChunk, annotations[j].githubAnnotation)
		}
		chunkedGitHubAnnotations = append(chunkedGitHubAnnotations, thisChunk)
	}

	check_title := fmt.Sprintf("Findings for %s", checkName)
	summary := fmt.Sprintf("A set of findings for %s on commit %s.", checkName, annotator.pr.headSHA)

	var first_annotations []*github.CheckRunAnnotation = nil
	if len(chunkedGitHubAnnotations) == 1 {
		first_annotations = chunkedGitHubAnnotations[0]
		chunkedGitHubAnnotations = nil
	} else if (len(chunkedGitHubAnnotations)) > 1 {
		first_annotations = chunkedGitHubAnnotations[0]
		chunkedGitHubAnnotations = chunkedGitHubAnnotations[1:]
	}

	output := github.CheckRunOutput{
		Title:   &check_title,
		Summary: &summary,

		Annotations: first_annotations,
	}

	conclusion := computeConclusion(annotations)
	completed_at := github.Timestamp{Time: time.Now()}
	options := github.CreateCheckRunOptions{
		Name:        checkName,
		HeadSHA:     annotator.pr.headSHA,
		Output:      &output,
		Conclusion:  &conclusion,
		CompletedAt: &completed_at,
	}
	checkRun, _, err := annotator.client.Checks.CreateCheckRun(context.Background(), annotator.pr.owner, annotator.pr.repo, options)
	if err != nil {
		return errors.Wrap(err, "failed to create check run")
	}

	for i, annotationChunk := range chunkedGitHubAnnotations {
		output := github.CheckRunOutput{
			Title:   &check_title, // required, even though there is no update
			Summary: &summary,     // required, even though there is no update

			Annotations: annotationChunk, // new annotations are appended (not overwritten)
		}
		options := github.UpdateCheckRunOptions{
			Name:   checkName, // required, even though there is no update
			Output: &output,
		}
		checkRun, _, err = annotator.client.Checks.UpdateCheckRun(context.Background(), annotator.pr.owner, annotator.pr.repo, checkRun.GetID(), options)
		if err != nil {
			return errors.Wrapf(err, "failed to post page %d of annotations - some posted successfully", i+2)
		}
	}

	return nil
}
