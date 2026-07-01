package scenarios

import (
	"github.com/eu-sovereign-cloud/conformance/pkg/builders"
	"github.com/eu-sovereign-cloud/conformance/pkg/generators"
	"github.com/eu-sovereign-cloud/conformance/pkg/mock/scenarios"
	"github.com/eu-sovereign-cloud/go-sdk/pkg/constants"
	"github.com/eu-sovereign-cloud/go-sdk/pkg/spec/schema"
	"github.com/eu-sovereign-cloud/terraform-provider-seca/internal/mock"
)

type ClientsInitParams struct {
	Region    string
	Zones     []string
	Providers []string
}

func ConfigureInitScenarioV1(scenario *scenarios.Scenario, params ClientsInitParams) error {
	configurator, err := scenario.StartConfiguration()
	if err != nil {
		return err
	}

	url := generators.GenerateRegionURL(constants.RegionProviderV1Name, params.Region)

	spec := &schema.RegionSpec{
		AvailableZones: []string{mock.ZoneA, mock.ZoneB},
		Providers:      builders.BuildProviderSpec(params.Providers, constants.ApiVersion1),
	}

	response, err := builders.NewRegionBuilder().
		Name(params.Region).
		Provider(constants.RegionProviderV1Name).ApiVersion(constants.ApiVersion1).
		Spec(spec).
		Build()
	if err != nil {
		return err
	}

	if err := configurator.ConfigureClientsInitStub(response, url, scenario.MockParams); err != nil {
		return err
	}

	if err := scenario.FinishConfiguration(configurator); err != nil {
		return err
	}
	return nil
}
