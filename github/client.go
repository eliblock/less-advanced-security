package github

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

type ClientConfiguration struct {
	AppID, InstallationID int64
	AppKeyPath            string
}

func createClient(configuration ClientConfiguration) (*github.Client, error) {
	itr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, configuration.AppID, configuration.InstallationID, configuration.AppKeyPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure GitHub access")
	}

	client := github.NewClient(&http.Client{Transport: itr})
	return client, nil
}
