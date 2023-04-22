package github

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

func TestLevelStringToNormalizedLevel(t *testing.T) {
	tests := []struct {
		in, outStr string
		out        int
	}{
		{"none", "notice", noticeLevel},
		{"note", "notice", noticeLevel},
		{"warning", "warning", warningLevel},
		{"error", "failure", failureLevel},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s returns %d", tt.in, tt.out), func(t *testing.T) {
			got, gotStr, err := levelStringToNormalizedLevel(tt.in)
			if err != nil {
				t.Errorf("expected no error but received %q", err)
			}
			if gotStr != tt.outStr {
				t.Errorf("expected %q, got %q", tt.outStr, gotStr)
			}
			if got != tt.out {
				t.Errorf("expected %d, got %d", tt.out, got)
			}
		})
	}
}

func TestAnnotationHash(t *testing.T) {
	tests := []struct {
		name       string
		annotation Annotation
		hash       string
	}{
		{
			"simple",
			Annotation{fileName: "test/file", startLine: 5, endLine: 5, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			"e7a950b00fee4a748050a65b913703b4",
		}, {
			"new start line",
			Annotation{fileName: "test/file", startLine: 4, endLine: 5, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			"5db942512fd238aef81561427f8a9114",
		}, {
			"new end line",
			Annotation{fileName: "test/file", startLine: 5, endLine: 6, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			"be58c0b17236dcbdce4f1f6adda85523",
		}, {
			"new message",
			Annotation{fileName: "test/file", startLine: 5, endLine: 5, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something somewhat bad")}},
			"3db8d84254a373e79261082972521974",
		}, {
			"new level",
			Annotation{fileName: "test/file", startLine: 5, endLine: 6, level: warningLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			"30ccf9386b64b10113362e0dd5e1ddbe",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmt.Sprintf("%x", tt.annotation.Hash()); got != tt.hash {
				t.Errorf("expected hash %s but received %s", tt.hash, got)
			}
		})
	}
}

func TestAnnotationAppendReportCount(t *testing.T) {
	tests := []struct {
		name          string
		annotation    Annotation
		reportCount   int
		expectedTitle string
	}{
		{
			"no change",
			Annotation{fileName: "test/file", startLine: 5, endLine: 5, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			1,
			"something bad",
		}, {
			"15 times",
			Annotation{fileName: "test/file", startLine: 5, endLine: 5, level: noticeLevel, githubAnnotation: &github.CheckRunAnnotation{Title: github.String("something bad")}},
			15,
			"something bad (reported 15 times)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.annotation.MaybeAppendReportCount(tt.reportCount)

			if tt.expectedTitle != *tt.annotation.githubAnnotation.Title {
				t.Errorf("expected annotation title to be %s but it was %s", tt.expectedTitle, *tt.annotation.githubAnnotation.Title)
			}
		})
	}
}

func TestLevelStringToNormalizedLevelErrors(t *testing.T) {
	tests := []string{"nil", "null", "info", "notice", "warn", "failure"}
	for _, tt_in := range tests {
		t.Run(tt_in, func(t *testing.T) {
			_, _, err := levelStringToNormalizedLevel(tt_in)
			if err == nil {
				t.Error("expected an error but received none")
			}
		})
	}
}

func compareAnnotation(t *testing.T, expected, received *Annotation) {
	t.Helper()

	if expected.fileName != received.fileName ||
		expected.level != received.level ||
		expected.startLine != received.startLine ||
		expected.endLine != received.endLine {
		t.Errorf("Annotation did not match expectation (expected %s, received %s)", expected, received)
	}

	expected_github_json, _ := json.Marshal(expected.githubAnnotation)
	received_github_json, _ := json.Marshal(received.githubAnnotation)
	if string(expected_github_json) != string(received_github_json) {
		t.Errorf("Github annotation did not match expectation: %s", cmp.Diff(expected_github_json, received_github_json))
	}
}
func TestCreateAnnotation(t *testing.T) {
	path := "src/main.py"
	startLine := 4
	endLine := 12
	level := "warning"
	title := "Finding title"
	message := "Finding message"
	details := "raw finding details"

	got, err := CreateAnnotation(path, startLine, endLine, level, title, message, details)
	if err != nil {
		t.Error(errors.Wrap(err, "failed to create annotation"))
	}

	compareAnnotation(t, got, &Annotation{
		githubAnnotation: &github.CheckRunAnnotation{
			Path:            &path,
			StartLine:       &startLine,
			EndLine:         &endLine,
			Title:           &title,
			Message:         &message,
			AnnotationLevel: &level,
			RawDetails:      &details,
		},
		level:     warningLevel,
		fileName:  path,
		startLine: startLine,
		endLine:   endLine,
	})

}

func TestRemoveEndLines(t *testing.T) {
	six := 6
	seven := 7
	twelve := 12

	one_line_annotation := Annotation{
		startLine:        6,
		endLine:          6,
		githubAnnotation: &github.CheckRunAnnotation{StartLine: &six, EndLine: &six},
	}
	multi_line_annotation := Annotation{
		startLine:        7,
		endLine:          12,
		githubAnnotation: &github.CheckRunAnnotation{StartLine: &seven, EndLine: &twelve},
	}
	flattened_multi_line_annotation := Annotation{
		startLine:        7,
		endLine:          7,
		githubAnnotation: &github.CheckRunAnnotation{StartLine: &seven, EndLine: &seven},
	}

	tests := []struct {
		name                             string
		annotations, expectedAnnotations []*Annotation
	}{
		{
			"one one-liner",
			[]*Annotation{&one_line_annotation},
			[]*Annotation{&one_line_annotation},
		},
		{
			"one multi-liner",
			[]*Annotation{&multi_line_annotation},
			[]*Annotation{&flattened_multi_line_annotation},
		},
		{
			"multiple",
			[]*Annotation{&one_line_annotation, &multi_line_annotation},
			[]*Annotation{&one_line_annotation, &flattened_multi_line_annotation},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			removeEndLines(tt.annotations)

			for _, expectedAnnotation := range tt.expectedAnnotations {
				found := false
				for _, gotAnnotation := range tt.annotations {
					if gotAnnotation.startLine == expectedAnnotation.startLine &&
						gotAnnotation.endLine == expectedAnnotation.endLine &&
						*gotAnnotation.githubAnnotation.StartLine == *expectedAnnotation.githubAnnotation.StartLine &&
						*gotAnnotation.githubAnnotation.EndLine == *expectedAnnotation.githubAnnotation.EndLine &&
						gotAnnotation.githubAnnotation.EndColumn == nil {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected annotation %s but did not find it.", expectedAnnotation)
				}
			}

			if len(tt.expectedAnnotations) != len(tt.annotations) {
				t.Errorf("expected %d annotations but got %d", len(tt.expectedAnnotations), len(tt.annotations))
			}
		})
	}
}
