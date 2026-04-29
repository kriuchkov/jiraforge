package app

import (
	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/adapters/atlassian"
	tempstore "github.com/kriuchkov/jiraforge/internal/adapters/storage/temp"
	appconfig "github.com/kriuchkov/jiraforge/internal/config"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
	servicejira "github.com/kriuchkov/jiraforge/internal/services/jira"
)

func LoadEnvFile(path string) error {
	if err := appconfig.LoadEnvFile(path); err != nil {
		return errors.Wrap(err, "load env file")
	}
	return nil
}

func LoadSettings() appconfig.Settings {
	return appconfig.LoadSettings()
}

func NewJiraService() (ports.JiraService, appconfig.Settings, error) {
	return newJiraService(LoadSettings())
}

func NewJiraServiceFromEnvFile(path string) (ports.JiraService, appconfig.Settings, error) {
	settings, err := appconfig.LoadSettingsFromEnvFile(path)
	if err != nil {
		return nil, settings, errors.Wrap(err, "load settings from env file")
	}
	return newJiraService(settings)
}

func newJiraService(settings appconfig.Settings) (ports.JiraService, appconfig.Settings, error) {
	if err := settings.Atlassian.Validate(); err != nil {
		return nil, settings, errors.Wrap(err, "validate Atlassian settings")
	}

	gateway, err := atlassian.NewGateway(settings.Atlassian)
	if err != nil {
		return nil, settings, errors.Wrap(err, "create Atlassian gateway")
	}

	service, err := servicejira.NewService(gateway, tempstore.NewStore())
	if err != nil {
		return nil, settings, errors.Wrap(err, "create Jira service")
	}

	return service, settings, nil
}
