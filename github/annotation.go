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
	fileName           string
	startLine, endLine int
	level              int
}

func levelStringToNormalizedLevel(level string) (normalizedLevel int, err error) {
	normalizedLevel = -1
	switch level {
	case "notice":
		normalizedLevel = noticeLevel
	case "warning":
		normalizedLevel = warningLevel
	case "failure":
		normalizedLevel = failureLevel
	}
	if normalizedLevel < 0 {
		err = errors.Errorf("invalid annotation level %v", level)
	}
	return
}

func CreateAnnotation(path string, startLine int, endLine int, level string, title string, message string, details string) (*Annotation, error) {
	normalizedLevel, err := levelStringToNormalizedLevel(level)
	if err != nil {
		return nil, errors.Wrap(err, "failed to normalize level")
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
		fileName:  path,
		startLine: startLine,
		endLine:   endLine,
	}, nil
}
