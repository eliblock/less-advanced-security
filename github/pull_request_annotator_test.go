package github

import (
	"fmt"
	"testing"
)

func TestComputeConclusion(t *testing.T) {
	noticeAnnotation := Annotation{level: noticeLevel}
	warningAnnotation := Annotation{level: warningLevel}
	failureAnnotation := Annotation{level: failureLevel}
	invalidAnnotation := Annotation{level: 4}

	tests := []struct {
		name        string
		annotations []*Annotation
		conclusion  string
	}{
		{"no annotations", []*Annotation{}, "success"},
		{"multiple notices", []*Annotation{&noticeAnnotation, &noticeAnnotation}, "success"},
		{"notice and warning", []*Annotation{&noticeAnnotation, &warningAnnotation}, "neutral"},
		{"notice and warning and failure", []*Annotation{&noticeAnnotation, &warningAnnotation, &failureAnnotation}, "failure"},
		{"notice and warning and failure reordered", []*Annotation{&failureAnnotation, &warningAnnotation, &noticeAnnotation}, "failure"},
		{"warning and invalid", []*Annotation{&invalidAnnotation, &warningAnnotation}, "neutral"},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s returns %s", tt.name, tt.conclusion), func(t *testing.T) {
			got := computeConclusion(tt.annotations)
			if tt.conclusion != got {
				t.Errorf("expected %q but got %q", tt.conclusion, got)
			}
		})
	}
}
