package main

import (
	"less-advanced-security/github"
	"less-advanced-security/sarif"

	"github.com/pkg/errors"
)

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

	return github.CreateAnnotation(result.Locations[0].Filepath, startLine, endLine, result.Level, title, result.Message, result.Raw)
}
