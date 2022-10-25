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

func (a Annotation) String() string {
	return fmt.Sprintf("{\"fileName\":%q,\"level\":%d,\"startLine\":%d,\"endLine\":%d,\"githubAnnotation\":\"...\"}", a.fileName, a.level, a.startLine, a.endLine)
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
