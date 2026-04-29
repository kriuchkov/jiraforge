package cli

import (
	"context"

	"github.com/go-faster/errors"
	"github.com/spf13/cobra"
)

type commandState struct {
	envFile string
	output  string
}

type commandRun func(context.Context) (any, string, error)

func (r runner) newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:              "jiraforge-cli",
		Short:            "JiraForge command line interface",
		Long:             "jiraforge-cli - JiraForge command line interface",
		Args:             cobra.NoArgs,
		SilenceErrors:    true,
		SilenceUsage:     true,
		TraverseChildren: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: func(cmd *cobra.Command, _ []string) error { _ = cmd.Help(); return errNoCommand },
	}

	state := &commandState{}

	// Set output for root command to ensure all subcommands inherit it
	root.SetOut(r.stderr)
	root.SetErr(r.stderr)

	// Global flags
	root.PersistentFlags().StringVar(&state.envFile, "env", "", "Path to .env file")
	root.PersistentFlags().StringVar(&state.output, "output", "text", "Output format: text or json")

	// Add subcommands
	root.AddCommand(
		r.newGetIssueCommand(state),
		r.newSearchIssuesCommand(state),
		r.newCreateIssueCommand(state),
		r.newCreateChildIssueCommand(state),
		r.newUpdateIssueCommand(state),
		r.newDeleteIssueCommand(state),
		r.newListIssueTypesCommand(state),
		r.newGetAgingReportCommand(state),
		r.newGetBlockedIssuesReportCommand(state),
		r.newGetFlowEfficiencyReportCommand(state),
		r.newGetSprintWorkloadReportCommand(state),
		r.newGetUtilizationReportCommand(state),
		r.newGetCycleTimeReportCommand(state),
		r.newGetTeamWipSnapshotCommand(state),
		r.newListSprintsCommand(state),
		r.newGetSprintCommand(state),
		r.newGetSprintReportCommand(state),
		r.newGetSprintHealthReportCommand(state),
		r.newGetActiveSprintCommand(state),
		r.newSearchSprintCommand(state),
		r.newAddCommentCommand(state),
		r.newGetCommentsCommand(state),
		r.newAddWorklogCommand(state),
		r.newGetWorklogsCommand(state),
		r.newGetTransitionsCommand(state),
		r.newTransitionIssueCommand(state),
		r.newListStatusesCommand(state),
		r.newGetIssueHistoryCommand(state),
		r.newGetRelatedIssuesCommand(state),
		r.newLinkIssuesCommand(state),
		r.newGetVersionCommand(state),
		r.newListProjectVersionsCommand(state),
		r.newGetDevelopmentInfoCommand(state),
		r.newDownloadAttachmentCommand(state),
	)
	return root
}

func (r runner) newCommand(state *commandState, use, short string, run commandRun) *cobra.Command {
	cmd := &cobra.Command{
		Use:           use,
		Short:         short,
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			value, text, err := run(cmd.Context())
			if err != nil {
				return errors.Wrap(err, use)
			}
			return r.writeResult(state.output, value, text)
		},
	}
	cmd.SetOut(r.stderr)
	cmd.SetErr(r.stderr)
	return cmd
}
