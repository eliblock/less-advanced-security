package github

import (
	"crypto/md5"
	"fmt"

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

func (a Annotation) String() string {
	checkRunAnnotationString := "{}"
	if a.githubAnnotation != nil {
		checkRunAnnotationString = checkRunAnnotationAsString(a.githubAnnotation)
	}
	return fmt.Sprintf("{\"fileName\":%q,\"level\":%d,\"startLine\":%d,\"endLine\":%d,\"githubAnnotation\":%s}", a.fileName, a.level, a.startLine, a.endLine, checkRunAnnotationString)
}

func (a Annotation) Hash() [16]byte {
	return md5.Sum([]byte(fmt.Sprintf("%d-%d-%s-%s-%d", a.startLine, a.endLine, a.fileName, *a.githubAnnotation.Title, a.level)))
}

func (a *Annotation) MaybeAppendReportCount(reportCount int) {
	/*
	* Call only once per annotation - if this is called multiple times, the report
	* count will be repeatedly appended.
	 */
	if reportCount > 1 {
		newTitle := fmt.Sprintf("%s (reported %d times)", *a.githubAnnotation.Title, reportCount)
		a.githubAnnotation.Title = &newTitle
	}
}

func checkRunAnnotationAsString(a *github.CheckRunAnnotation) string {
	return fmt.Sprintf("{\"path\":%s,\"message\":%s,\"title\":%s,\"...\":\"...\"}", *a.Path, *a.Message, *a.Title)
}

// map sarif levels (https://docs.oasis-open.org/sarif/sarif/v2.0/sarif-v2.0.html#_Ref508894469) to GitHub levels
func levelStringToNormalizedLevel(level string) (normalizedLevel int, normalizedLevelString string, err error) {
	normalizedLevel = -1
	normalizedLevelString = ""
	switch level {
	case "none":
		normalizedLevel = noticeLevel
		normalizedLevelString = "notice"
	case "note":
		normalizedLevel = noticeLevel
		normalizedLevelString = "notice"
	case "warning":
		normalizedLevel = warningLevel
		normalizedLevelString = "warning"
	case "error":
		normalizedLevel = failureLevel
		normalizedLevelString = "failure"
	}
	if normalizedLevel < 0 {
		err = errors.Errorf("invalid annotation level %v", level)
	}
	return
}

func CreateAnnotation(path string, startLine int, endLine int, level string, title string, message string) (*Annotation, error) {
	normalizedLevel, normalizedLevelString, err := levelStringToNormalizedLevel(level)
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
			AnnotationLevel: &normalizedLevelString,
		},
		level:     normalizedLevel,
		fileName:  path,
		startLine: startLine,
		endLine:   endLine,
	}, nil
}

func removeEndLines(annotations []*Annotation) {
	for _, annotation := range annotations {
		annotation.endLine = annotation.startLine
		annotation.githubAnnotation.EndLine = annotation.githubAnnotation.StartLine
		annotation.githubAnnotation.EndColumn = nil
	}
}
