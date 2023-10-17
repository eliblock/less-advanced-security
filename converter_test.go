package main

import (
	"reflect"
	"testing"

	"less-advanced-security/github"
	"less-advanced-security/sarif"
)

func TestSarifToAnnotationConverter(t *testing.T) {
	five, ten := 5, 10

	sarifWithStartLine := sarif.Result{
		Message: "this is a failure",
		RuleID:  "fail-1-2-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationWithStartLine, _ := github.CreateAnnotation("test/file", five, five, "error", "fail-1-2-3", "this is a failure")

	sarifWithStartAndEndLine := sarif.Result{
		Message: "this is a failure",
		RuleID:  "fail-1-2-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five, EndLine: &ten}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationWithStartAndEndLine, _ := github.CreateAnnotation("test/file", five, ten, "error", "fail-1-2-3", "this is a failure")

	tests := []struct {
		name       string
		result     sarif.Result
		annotation *github.Annotation
		errMessage string
	}{
		{
			"no locations",
			sarif.Result{Locations: []sarif.ResultLocation{}},
			&github.Annotation{},
			"each result must have 1 location, not 0",
		},
		{
			"two locations",
			sarif.Result{Locations: []sarif.ResultLocation{sarif.ResultLocation{}, sarif.ResultLocation{}}},
			&github.Annotation{},
			"each result must have 1 location, not 2",
		},
		{
			"no start line",
			sarif.Result{Locations: []sarif.ResultLocation{sarif.ResultLocation{Filepath: "test/file"}}},
			&github.Annotation{},
			"each result must have a start line",
		},
		{
			"start line only",
			sarifWithStartLine,
			annotationWithStartLine,
			"",
		},
		{
			"start and end line",
			sarifWithStartAndEndLine,
			annotationWithStartAndEndLine,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resultToAnnotation(tt.result)
			if tt.errMessage != "" {
				if err == nil || err.Error() != tt.errMessage {
					t.Errorf("Expected error %q but got %q.", tt.errMessage, err)
				}
			} else if err != nil {
				t.Errorf("Got error %q but expected no error.", err)
			} else if !reflect.DeepEqual(got, tt.annotation) {
				t.Errorf("Expected annotation %s but got %s.", tt.annotation, got)
			}
		})
	}
}

func TestSarifsToAnnotationsConverter(t *testing.T) {
	five, six, ten := 5, 6, 10

	sarifWithNoLocation := sarif.Result{Locations: []sarif.ResultLocation{}}

	sarifOriginal := sarif.Result{
		Message: "this is a failure",
		RuleID:  "fail-1-2-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationOriginal, _ := github.CreateAnnotation("test/file", five, five, "error", "fail-1-2-3", "this is a failure")
	annotationOriginalReportedTwice, _ := github.CreateAnnotation("test/file", five, five, "error", "fail-1-2-3 (reported 2 times)", "this is a failure")

	sarifAsWarning := sarif.Result{
		Message: "this is a failure",
		RuleID:  "fail-1-2-3",
		Raw:     "raw failure text",
		Level:   "warning",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationAsWarning, _ := github.CreateAnnotation("test/file", five, five, "warning", "fail-1-2-3", "this is a failure")
	annotationAsWarningReportedTwice, _ := github.CreateAnnotation("test/file", five, five, "warning", "fail-1-2-3 (reported 2 times)", "this is a failure")

	sarifNewId := sarif.Result{
		Message: "this is a failure",
		RuleID:  "new-id-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationNewId, _ := github.CreateAnnotation("test/file", five, five, "error", "new-id-3", "this is a failure")

	sarifNewStartLine := sarif.Result{
		Message: "this is a failure",
		RuleID:  "new-id-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &six}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationNewStartLine, _ := github.CreateAnnotation("test/file", six, six, "error", "new-id-3", "this is a failure")

	sarifNewEndLine := sarif.Result{
		Message: "this is a failure",
		RuleID:  "new-id-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five, EndLine: &ten}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationNewEndLine, _ := github.CreateAnnotation("test/file", five, ten, "error", "new-id-3", "this is a failure")

	tests := []struct {
		name                string
		results             []*sarif.Result
		expectedAnnotations []*github.Annotation
		errMessage          string
	}{
		{
			"no locations",
			[]*sarif.Result{&sarifWithNoLocation},
			[]*github.Annotation{},
			"failed to normalize result: each result must have 1 location, not 0",
		}, {
			"two results",
			[]*sarif.Result{&sarifOriginal, &sarifAsWarning},
			[]*github.Annotation{annotationOriginal, annotationAsWarning},
			"",
		}, {
			"two sets of duplicate results",
			[]*sarif.Result{&sarifOriginal, &sarifAsWarning, &sarifOriginal, &sarifAsWarning},
			[]*github.Annotation{annotationOriginalReportedTwice, annotationAsWarningReportedTwice},
			"",
		}, {
			"not duplicated due to start line",
			[]*sarif.Result{&sarifOriginal, &sarifNewStartLine},
			[]*github.Annotation{annotationOriginal, annotationNewStartLine},
			"",
		}, {
			"not duplicated due to end line",
			[]*sarif.Result{&sarifOriginal, &sarifNewEndLine},
			[]*github.Annotation{annotationOriginal, annotationNewEndLine},
			"",
		}, {
			"not duplicated due to id",
			[]*sarif.Result{&sarifOriginal, &sarifNewId},
			[]*github.Annotation{annotationOriginal, annotationNewId},
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAnnotations, err := resultsToAnnotations(tt.results)
			if tt.errMessage != "" {
				if err == nil || err.Error() != tt.errMessage {
					t.Errorf("Expected error %q but got %q.", tt.errMessage, err)
				}
			} else if err != nil {
				t.Errorf("Expected no error but got %q.", err)
			} else {
				for _, expectedAnnotation := range tt.expectedAnnotations {
					found := false
					for _, gotAnnotation := range gotAnnotations {
						if reflect.DeepEqual(gotAnnotation, expectedAnnotation) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected annotation %s but did not find it.", expectedAnnotation)
					}
				}

				if len(tt.expectedAnnotations) != len(gotAnnotations) {
					t.Errorf("expected %d annotations but got %d", len(tt.expectedAnnotations), len(gotAnnotations))
				}
			}
		})
	}
}
