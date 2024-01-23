package main

import (
	"flag"
	"fmt"
	"less-advanced-security/github"
	"less-advanced-security/sarif"
	"log"
	"os"
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
	appKeyPath := flag.String(
		"key_path", "",
		"absolute path to your GitHub app's private key. "+
			"The environment variable APP_KEY can also be used instead (the flag takes precedence).",
	)

	sarifPath := flag.String("sarif_path", "", "absolute path to your sarif file")
	checkNameOverride := flag.String("check_name", "", "name of the check, defaults to tool name from sarif")

	filterAnnotations := flag.Bool("filter_annotations", true, "filter annotations by lines found in the git patches, default true")
	annotateStartLineOnly := flag.Bool("annotate_beginning", true, "force annotations to start line of a finding (if set to false, GitHub default of end is used), default true")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("%s\n", version)
		return
	}

	var err error

	// Optionally loads the application's private key from the APP_KEY environment variable.
	// If it's not defined, then it will be blank and we will attempt to load a file instead (if provided).
	appKey := []byte(os.Getenv("APP_KEY"))

	// Load the GitHub application's private key from path if provided.
	// This flag takes precedence over the APP_KEY variable.
	if *appKeyPath != "" {
		if appKey, err = os.ReadFile(*appKeyPath); err != nil {
			log.Fatal(errors.Wrap(err, "failed to load the GitHub app's private key"))
		}
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
		github.ClientConfiguration{AppID: int64(*appID), InstallationID: int64(*installID), AppKey: appKey},
		github.PullRequestConfiguration{Owner: parsedRepo[0], Repo: parsedRepo[1], Number: *prNumber},
		*sha,
	)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed during setup"))
	}

	annotations, err := resultsToAnnotations(results)
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to convert results to annotations"))
	}

	checkName := tool.Name
	if *checkNameOverride != "" {
		checkName = *checkNameOverride
	}

	if err := annotator.PostAnnotations(annotations, checkName, *filterAnnotations, *annotateStartLineOnly); err != nil {
		log.Fatal(errors.Wrap(err, "failed to post annotations"))
	}
}
