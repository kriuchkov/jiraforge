package main

import (
	"context"
	"os"

	"github.com/go-faster/errors"

	cliadapter "github.com/kriuchkov/jiraforge/internal/adapters/commands/cli"
	"github.com/kriuchkov/jiraforge/internal/app"
	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

func main() {
	exitCode := cliadapter.Execute(context.Background(), os.Args[1:], os.Stdout, os.Stderr, func(envFile string) (ports.JiraService, error) {
		service, _, err := app.NewJiraServiceFromEnvFile(envFile)
		if err != nil {
			return nil, errors.Wrap(err, "create Jira service")
		}
		return service, nil
	})
	os.Exit(exitCode)
}
