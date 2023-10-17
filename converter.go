package main

import (
	"less-advanced-security/github"
	"less-advanced-security/sarif"

	"github.com/pkg/errors"
)

type AnnotationCount struct {
	annotation *github.Annotation
	count      int
}

func resultsToAnnotations(results []*sarif.Result) ([]*github.Annotation, error) {
	annotationHashToCount := make(map[[16]byte]*AnnotationCount)
	for _, result := range results {
		if result == nil {
			continue
		}
		annotation, err := resultToAnnotation(*result)
		if err != nil {
			return nil, errors.Wrap(err, "failed to normalize result")
		}

		annotationHash := annotation.Hash()
		if annotationHashToCount[annotationHash] != nil {
			annotationHashToCount[annotationHash].count += 1
		} else {
			annotationHashToCount[annotationHash] = &AnnotationCount{annotation: annotation, count: 1}
		}
	}
	annotations := []*github.Annotation{}
	for _, annotationCount := range annotationHashToCount {
		annotationCount.annotation.MaybeAppendReportCount(annotationCount.count)
		annotations = append(annotations, annotationCount.annotation)
	}

	return annotations, nil
}

func resultToAnnotation(result sarif.Result) (*github.Annotation, error) {
	if len(result.Locations) != 1 {
		return nil, errors.Errorf("each result must have 1 location, not %d", len(result.Locations))
	}
	if result.Locations[0].StartLine == nil {
		return nil, errors.Errorf("each result must have a start line")
	}
	startLine := *result.Locations[0].StartLine

	endLine := startLine
	if result.Locations[0].EndLine != nil {
		endLine = *result.Locations[0].EndLine
	}

	title := result.RuleID

	return github.CreateAnnotation(result.Locations[0].Filepath, startLine, endLine, result.Level, title, result.Message)
}
