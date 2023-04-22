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
	annotationWithStartLine, _ := github.CreateAnnotation("test/file", five, five, "error", "fail-1-2-3", "this is a failure", "raw failure text")

	sarifWithStartAndEndLine := sarif.Result{
		Message: "this is a failure",
		RuleID:  "fail-1-2-3",
		Raw:     "raw failure text",
		Level:   "error",
		Locations: []sarif.ResultLocation{
			sarif.ResultLocation{Filepath: "test/file", StartLine: &five, EndLine: &ten}},
	}
	// accuracy of annotation creation tested elsewhere
	annotationWithStartAndEndLine, _ := github.CreateAnnotation("test/file", five, ten, "error", "fail-1-2-3", "this is a failure", "raw failure text")

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
