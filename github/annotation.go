package github

import (
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

const (
	failureLevel = iota
	warningLevel
	noticeLevel
)

type Annotation struct {
	githubAnnotation   *github.CheckRunAnnotation
	filePath           string
	startLine, endLine int
	level              int
}

func CreateAnnotation(path string, startLine int, endLine int, level string, title string, message string, details string) (*Annotation, error) {
	normalizedLevel := -1
	switch level {
	case "notice":
		normalizedLevel = noticeLevel
		break
	case "warning":
		normalizedLevel = warningLevel
		break
	case "failure":
		normalizedLevel = failureLevel
		break
	}
	if normalizedLevel < 0 {
		return nil, errors.Errorf("invalid annotation level %v", level)
	}

	return &Annotation{
		githubAnnotation: &github.CheckRunAnnotation{
			Path:            &path,
			StartLine:       &startLine,
			EndLine:         &endLine,
			Title:           &title,
			Message:         &message,
			AnnotationLevel: &level,
			RawDetails:      &details,
		},
		level:     normalizedLevel,
		filePath:  path,
		startLine: startLine,
		endLine:   endLine,
	}, nil
}
