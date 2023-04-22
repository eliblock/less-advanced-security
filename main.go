package main

import (
	"flag"
	"fmt"
	"less-advanced-security/github"
	"less-advanced-security/sarif"
	"log"
	"strings"

	"github.com/pkg/errors"
)

var (
	version = "dev"
)

func main() {
	versionFlag := flag.Bool("version", false, "")

	repo := flag.String("repo", "", "repo in the form ownerName/repoName")
	sha := flag.String("sha", "", "SHA of the commit to annotate")
	prNumber := flag.Int("pr", -1, "id of pr to annotate")

	appID := flag.Int("app_id", -1, "app id for your GitHub app")
	installID := flag.Int("install_id", -1, "install id for your GitHub app installation")
	appKeyPath := flag.String("key_path", "", "absolute path to your GitHub app's private key")

	sarifPath := flag.String("sarif_path", "", "absolute path to your sarif file")
	checkNameOverride := flag.String("check_name", "", "name of the check, defaults to tool name from sarif")

	filterAnnotations := flag.Bool("filter_annotations", true, "filter annotations by lines found in the git patches, default true")
	annotateStartLineOnly := flag.Bool("annotate_beginning", true, "force annotations to start line of a finding (if set to false, GitHub default of end is used), default true")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", version)
		return
	}

	parsedRepo := strings.Split(*repo, "/")

	tool, results, err := sarif.ParseFromFile(*sarifPath)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to load sarif file"))
	}

	if len(results) == 0 {
		log.Println("No findings to post.")
		return
	}

	annotator, err := github.CreatePullRequestAnnotator(
		github.ClientConfiguration{AppID: int64(*appID), InstallationID: int64(*installID), AppKeyPath: *appKeyPath},
		github.PullRequestConfiguration{Owner: parsedRepo[0], Repo: parsedRepo[1], Number: *prNumber},
		*sha,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed during setup"))
	}

	annotations := []*github.Annotation{}
	for _, result := range results {
		if result == nil {
			continue
		}
		annotation, err := resultToAnnotation(*result)
		if err != nil {
			log.Fatal(errors.Wrap(err, "failed to normalize result"))
		}
		annotations = append(annotations, annotation)
	}

	checkName := tool.Name
	if *checkNameOverride != "" {
		checkName = *checkNameOverride
	}

	if err := annotator.PostAnnotations(annotations, checkName, *filterAnnotations, *annotateStartLineOnly); err != nil {
		log.Fatal(errors.Wrap(err, "failed to post annotations"))
	}
}
