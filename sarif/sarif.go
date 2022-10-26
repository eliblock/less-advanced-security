package sarif

import (
	"encoding/json"
	"fmt"

	"github.com/owenrumney/go-sarif/sarif"
	"github.com/pkg/errors"
)

type Tool struct {
	Name    string
	Version *string
}

type Result struct {
	Message   string
	RuleID    string
	Locations []ResultLocation
	Raw       string
	Level     string
}

type ResultLocation struct {
	Filepath           string
	StartLine, EndLine *int
}

func ParseFromFile(path string) (*Tool, []*Result, error) {
	report, err := sarif.Open(path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to load sarif file")
	}

	if len(report.Runs) > 1 {
		return nil, nil, errors.Errorf("cannot parse more than 1 run, found %d", len(report.Runs))
	}

	run := report.Runs[0]

	tool := Tool{
		Name:    run.Tool.Driver.Name,
		Version: run.Tool.Driver.Version,
	}

	ruleid_to_level := map[string]string{}
	for _, rule := range run.Tool.Driver.Rules {
		if rule.DefaultConfiguration != nil && rule.DefaultConfiguration.Level != nil {
			ruleid_to_level[rule.ID] = fmt.Sprintf("%v", rule.DefaultConfiguration.Level)
		}
	}

	results := []*Result{}
	for _, result := range run.Results {
		raw, _ := json.Marshal(result)

		var locations []ResultLocation
		for _, location := range result.Locations {
			if location.PhysicalLocation == nil || location.PhysicalLocation.ArtifactLocation == nil {
				continue
			}

			var startLine, endLine *int
			if location.PhysicalLocation.Region != nil {
				startLine = location.PhysicalLocation.Region.StartLine
				endLine = location.PhysicalLocation.Region.EndLine
			}
			locations = append(locations, ResultLocation{
				Filepath:  *location.PhysicalLocation.ArtifactLocation.URI,
				StartLine: startLine,
				EndLine:   endLine,
			})
		}

		var level string
		if result.Level != nil {
			level = *result.Level
		} else {
			level = ruleid_to_level[*result.RuleID]
		}

		results = append(results, &Result{
			Message:   *result.Message.Text,
			RuleID:    *result.RuleID,
			Raw:       string(raw),
			Locations: locations,
			Level:     level,
		})

	}

	return &tool, results, nil
}
