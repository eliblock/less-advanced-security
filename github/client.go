package github

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v47/github"
	"github.com/pkg/errors"
)

type ClientConfiguration struct {
	AppID, InstallationID int64
	AppKey                []byte
}

func createClient(configuration ClientConfiguration) (*github.Client, error) {
	itr, err := ghinstallation.New(http.DefaultTransport, configuration.AppID, configuration.InstallationID, configuration.AppKey)
	if err != nil {
		return nil, errors.Wrap(err, "failed to configure GitHub access")
	}

	client := github.NewClient(&http.Client{Transport: itr})
	return client, nil
}
