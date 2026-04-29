package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteMetaCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		args            []string
		wantExitCode    int
		wantInStderr    []string
		wantEmptyStdout bool
	}{
		{
			name:            "no args prints usage",
			args:            nil,
			wantExitCode:    1,
			wantInStderr:    []string{"Usage:", "jiraforge-cli - JiraForge command line interface"},
			wantEmptyStdout: true,
		},
		{
			name:            "help prints usage",
			args:            []string{"help"},
			wantExitCode:    0,
			wantInStderr:    []string{"Usage:", "Available Commands:"},
			wantEmptyStdout: true,
		},
		{
			name:            "unknown command prints usage",
			args:            []string{"nope"},
			wantExitCode:    1,
			wantInStderr:    []string{"unknown command: nope", "jiraforge-cli - JiraForge command line interface"},
			wantEmptyStdout: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := &stubCLIService{}
			stdout, stderr, exitCode := executeCLI(t, tt.args, service, newStubFactory(service))

			require.Equal(t, tt.wantExitCode, exitCode)
			if tt.wantEmptyStdout {
				require.Empty(t, stdout)
			}
			for _, want := range tt.wantInStderr {
				require.Contains(t, stderr, want)
			}
		})
	}
}
