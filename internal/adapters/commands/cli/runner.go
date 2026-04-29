package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/go-faster/errors"

	"github.com/kriuchkov/jiraforge/internal/core/ports"
)

var errNoCommand = errors.New("no command provided")

type ServiceFactory func(envFile string) (ports.JiraService, error)

type runner struct {
	stdout  io.Writer
	stderr  io.Writer
	factory ServiceFactory
}

func Execute(ctx context.Context, args []string, stdout, stderr io.Writer, factory ServiceFactory) int {
	r := runner{stdout: stdout, stderr: stderr, factory: factory}
	root := r.newRootCommand()
	root.SetArgs(args)

	if len(args) == 0 {
		_ = root.Help()
		return 1
	}

	err := root.ExecuteContext(ctx)
	switch {
	case err == nil:
		return 0
	case errors.Is(err, errNoCommand):
		return 1
	case isUnknownCommand(err) && len(args) > 0:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		_ = root.Help()
		return 1
	default:
		fmt.Fprintln(stderr, err.Error())
		return 1
	}
}

func isUnknownCommand(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "unknown command ")
}

func (r runner) service(envFile string) (ports.JiraService, error) {
	return r.factory(envFile)
}

func (r runner) writeResult(output string, value any, text string) error {
	if output != "text" && output != "json" {
		return fmt.Errorf("unsupported output format %q: use text or json", output)
	}
	if output == "json" {
		encoder := json.NewEncoder(r.stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(value); err != nil {
			return errors.Wrap(err, "write JSON output")
		}
		return nil
	}
	_, err := fmt.Fprintln(r.stdout, text)
	if err != nil {
		return errors.Wrap(err, "write text output")
	}
	return nil
}

func splitCSV(value string) []string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	parts := strings.Split(trimmed, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
