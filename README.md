# JiraForge

<p align="center">
  <a href="https://go.dev">
    <img src="https://img.shields.io/badge/Built%20with-Go-00ADD8?style=flat-square&logo=go&logoColor=white" 
         alt="Built with Go">
  </a>

  <a href="https://github.com/kriuchkov/jiraforge/blob/main/LICENSE">
    <img src="https://img.shields.io/github/license/kriuchkov/jiraforge?style=flat-square&color=blue" 
        alt="License">
  </a>

  <a href="https://github.com/kriuchkov/jiraforge/stargazers">
    <img src="https://img.shields.io/github/stars/kriuchkov/jiraforge?style=flat-square" 
         alt="Stars">
  </a>

  <a href="https://www.atlassian.com/software/jira">
    <img src="https://img.shields.io/badge/Integration-Jira-0052CC?style=flat-square&logo=jira&logoColor=white" 
         alt="Jira Integration">
  </a>

  <a href="https://modelcontextprotocol.io">
    <img src="https://img.shields.io/badge/Protocol-MCP-orange?style=flat-square" 
    alt="MCP Protocol">
  </a>

  <a href="https://cursor.com/">
      <img src="https://img.shields.io/badge/Compatible%20with-Cursor-5E5E5E?style=flat-square"
       alt="Compatible with Cursor">
  </a>

  <a href="https://code.visualstudio.com/">
      <img src="https://img.shields.io/badge/Compatible%20with-VS%20Code-007ACC?style=flat-square&logo=visual-studio-code&logoColor=white" alt="Compatible with VS Code">
  </a>
</p>

JiraForge is a Jira operations layer for teams that want Jira available in AI clients, terminal workflows, and local automation without re-implementing the same integration three times.

At a product level, JiraForge helps turn Jira from a browser-only workflow into a reusable tool surface. Instead of constantly switching between Jira UI, custom scripts, and AI tools, a team gets one consistent way to read issues, comments, sprints, statuses, versions, linked work, and development context, and to perform common write operations when needed.

This is useful when you want to move faster on operational Jira work such as triage, release readiness checks, sprint inspection, issue follow-up, and development context gathering. It is especially helpful for teams that already work in Cursor, VS Code, Claude, shells, or local agent workflows and want Jira to be part of that same working environment.

## What JiraForge helps with

- faster issue triage without manually clicking through Jira screens
- release and version reviews that combine Jira issue state with development signals
- sprint and workflow inspection from chat, CLI, or MCP tools
- issue investigation using comments, history, transitions, related issues, and development metadata in one place
- repeatable Jira operations from scripts, local tooling, and AI-assisted workflows

The project exposes one shared Jira service through three entrypoints:

- `jiraforge`: MCP server for Claude, Cursor, VS Code, and other MCP clients
- `jiraforge-cli`: terminal-first interface for direct Jira operations
- `jiraforge-agent`: Gemini agent built with ADK Go, connected to the same MCP tool surface

## When to use each binary

Use `jiraforge` when you want JiraForge to act as an MCP server for another client. This is the right choice for Claude Desktop, Cursor, VS Code MCP, or any other tool that already knows how to talk to an MCP endpoint.

Use `jiraforge-cli` when you want explicit, repeatable terminal commands. It is the best fit for shell usage, automation, CI jobs, and cases where you want predictable inputs and machine-readable JSON output.

Use `jiraforge-agent` when you want a chat-driven Jira assistant on top of the same local MCP tools. It is useful for multi-step requests such as summarizing an issue, checking related work and development state, or deciding on the next Jira action from a natural-language prompt. It does not add a separate Jira implementation; it adds Gemini-based reasoning and tool orchestration, and it requires `GOOGLE_API_KEY`.
It also supports persistent chat sessions, so you can resume the same conversation later by user and session ID instead of starting from an empty context every time.

## Requirements

- Atlassian Cloud host, email, and API token
- `GOOGLE_API_KEY` only if you want to run the Gemini ADK agent

## Install

Install from Homebrew:

```bash
brew tap kriuchkov/homebrew-tap
brew install jiraforge
brew install jiraforge-cli
brew install jiraforge-agent
```

If you prefer local builds from source, use the commands in the Build section below.

## Configuration

Required Jira variables:

```bash
ATLASSIAN_HOST=https://your-company.atlassian.net
ATLASSIAN_EMAIL=your-email@company.com
ATLASSIAN_TOKEN=your-api-token
```

Optional Gemini variables:

```bash
GOOGLE_API_KEY=your-google-ai-api-key
GEMINI_MODEL=gemini-2.5-flash
```

You can export them in the shell or place them in a local `.env` file and pass `--env .env` to the binaries.

## Build

```bash
make build
make build-cli
make build-agent
```

Or directly:

```bash
go build -o bin/jiraforge ./cmd/jiraforge
go build -o bin/jiraforge-cli ./cmd/jiraforge-cli
go build -o bin/jiraforge-agent ./cmd/jiraforge-agent
```

## Basic usage

The examples in this section assume you built the binaries locally into `./bin`. If you installed with Homebrew, use the same commands without the `./bin/` prefix.

### `jiraforge`

Use the MCP server when you want Claude, Cursor, VS Code, or another MCP client to talk to Jira through JiraForge.

Start it in stdio mode when the MCP client will launch the process itself:

```bash
./bin/jiraforge --env .env
```

Start it in HTTP mode when you want to debug locally or connect to `http://localhost:3000/mcp`:

```bash
./bin/jiraforge --env .env --http_port 3000
```

### `jiraforge-cli`

Use the CLI when you want direct terminal access to Jira operations without an MCP client or LLM agent.

See available commands:

```bash
./bin/jiraforge-cli --help
```

Read an issue:

```bash
./bin/jiraforge-cli get-issue --env .env --issue-key PROJ-123
```

Search with JQL and return JSON for scripts:

```bash
./bin/jiraforge-cli search-issues --env .env --jql "assignee = currentUser()" --output json
```

### `jiraforge-agent`

Use the Gemini agent when you want a chat-driven interface on top of the same local JiraForge MCP tools. In addition to the Jira variables, this entrypoint also requires `GOOGLE_API_KEY`.

Start the agent with the default model:

```bash
./bin/jiraforge-agent console --env .env
```

Resume the most recent session for the same user:

```bash
./bin/jiraforge-agent console --env .env --user-id alice --resume-last
```

Pin work to an explicit session ID:

```bash
./bin/jiraforge-agent console --env .env --user-id alice --session-id sprint-triage
```

Manage saved sessions without starting Gemini or connecting to Jira:

```bash
./bin/jiraforge-agent sessions list --user-id alice
./bin/jiraforge-agent sessions inspect --user-id alice --session-id sprint-triage
./bin/jiraforge-agent sessions delete --user-id alice --session-id sprint-triage
```

Override the Gemini model explicitly:

```bash
./bin/jiraforge-agent console --env .env --model gemini-2.5-flash
```

## Atlassian adapter integration tests

These tests hit a real Jira instance directly against the `internal/adapters/atlassian` package. They do not bootstrap MCP, CLI, or the full application runtime.

Required environment for read-only scenarios:

```bash
export ATLASSIAN_HOST=https://your-company.atlassian.net
export ATLASSIAN_EMAIL=your-email@company.com
export ATLASSIAN_TOKEN=your-api-token
export JIRAFORGE_TEST_ISSUE_KEY=PROJ-123
```

Optional environment for broader coverage:

```bash
export JIRAFORGE_TEST_PROJECT_KEY=PROJ
export JIRAFORGE_TEST_BOARD_ID=123
export JIRAFORGE_TEST_ATTACHMENT_ID=456789
export JIRAFORGE_TEST_ENABLE_MUTATIONS=1
export JIRAFORGE_TEST_MUTATION_PROJECT_KEY=PROJ
export JIRAFORGE_TEST_CREATE_ISSUE_TYPE=Task
export JIRAFORGE_TEST_LINK_TARGET_ISSUE_KEY=PROJ-456
export JIRAFORGE_TEST_LINK_TYPE=Relates
```

Run only this layer:

```bash
make test-atlassian-integration
```

Or directly:

```bash
go test -count=1 -tags=integration ./internal/adapters/atlassian -v
```

Read-only coverage runs with the required variables. Mutation scenarios are skipped unless `JIRAFORGE_TEST_ENABLE_MUTATIONS=1` is set.

There is also a manual GitHub Actions workflow for the same layer in `.github/workflows/atlassian-integration.yaml`.
Store `ATLASSIAN_HOST`, `ATLASSIAN_EMAIL`, and `ATLASSIAN_TOKEN` as repository secrets, then run `Atlassian Adapter Integration Tests` from the Actions tab and fill the workflow inputs that map to the `JIRAFORGE_TEST_*` variables.

## MCP server

Run in stdio mode:

```bash
jiraforge --env .env
```

Run in HTTP mode for local debugging:

```bash
jiraforge --env .env --http_port 3000
```

Cursor or Claude Desktop MCP configuration for stdio mode:

```json
{
  "mcpServers": {
    "jira": {
      "command": "/absolute/path/to/jiraforge",
      "args": ["--env", "/absolute/path/to/.env"]
    }
  }
}
```

Cursor MCP configuration for HTTP mode:

```json
{
  "mcpServers": {
    "jira": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

### MCP tools

Read operations:

- `jira_get_issue`
- `jira_search_issue`
- `jira_list_issue_types`
- `jira_list_sprints`
- `jira_get_aging_report`
- `jira_get_blocked_issues_report`
- `jira_get_flow_efficiency_report`
- `jira_get_sprint_workload_report`
- `jira_get_utilization_report`
- `jira_get_cycle_time_report`
- `jira_get_team_wip_snapshot`
- `jira_get_worklogs`
- `jira_get_sprint`
- `jira_get_sprint_report`
- `jira_get_sprint_health_report`
- `jira_get_active_sprint`
- `jira_search_sprint_by_name`
- `jira_get_comments`
- `jira_get_transitions`
- `jira_list_statuses`
- `jira_get_issue_history`
- `jira_get_related_issues`
- `jira_get_version`
- `jira_list_project_versions`
- `jira_get_development_information`
- `jira_download_attachment`

Write operations:

- `jira_create_issue`
- `jira_create_child_issue`
- `jira_update_issue`
- `jira_delete_issue`
- `jira_add_comment`
- `jira_add_worklog`
- `jira_transition_issue`
- `jira_link_issues`

Available prompts:

- `issue_development_tree`
- `release_development_overview`

## CLI

Run help:

```bash
jiraforge-cli --help
```

Every command supports:

- `--env` for `.env` loading
- `--output text|json` for human-readable or machine-readable output (applies to every command, not just `search-issues`)

Examples:

```bash
# issue inspection
jiraforge-cli get-issue --env .env --issue-key PROJ-123

# search with JQL
jiraforge-cli search-issues --env .env --jql "project = PROJ ORDER BY updated DESC" --max-results 20

# create an issue
jiraforge-cli create-issue --env .env \
  --project-key PROJ \
  --summary "Fix login redirect" \
  --description "Users are redirected to the wrong page after login." \
  --issue-type Bug

# add a comment
jiraforge-cli add-comment --env .env --issue-key PROJ-123 --comment "Investigating the regression."

# transition an issue
jiraforge-cli transition-issue --env .env --issue-key PROJ-123 --transition-id 31

# inspect linked development state
jiraforge-cli get-development-info --env .env --issue-key PROJ-123

# inspect a sprint report
jiraforge-cli get-sprint-report --env .env --sprint-id 42

# inspect sprint health
jiraforge-cli get-sprint-health-report --env .env --sprint-id 42

# find stale work in active statuses
jiraforge-cli get-aging-report --env .env --project-key PROJ --status "In Progress" --status "Code Review" --min-days-in-status 5

# inspect blocked work and dependencies
jiraforge-cli get-blocked-issues-report --env .env --project-key PROJ --status "In Progress" --status "Blocked" --blocked-status "Blocked"

# inspect weekly active-vs-blocked flow efficiency with assignee rollups
jiraforge-cli get-flow-efficiency-report --env .env --project-key PROJ --status "In Progress" --status "Code Review" --status "Blocked" --active-status "In Progress" --active-status "Code Review" --blocked-status "Blocked" --window-days 7

# inspect an explicit review period instead of a rolling window
jiraforge-cli get-flow-efficiency-report --env .env --project-key PROJ --status "In Progress" --status "Code Review" --status "Blocked" --active-status "In Progress" --active-status "Code Review" --blocked-status "Blocked" --start-date 2026-04-01 --end-date 2026-04-07

# machine-readable output
jiraforge-cli search-issues --env .env --jql "assignee = currentUser()" --output json
```

CLI commands:

- `get-issue`
- `search-issues`
- `create-issue`
- `create-child-issue`
- `update-issue`
- `delete-issue`
- `list-issue-types`
- `get-aging-report`
- `get-blocked-issues-report`
- `get-flow-efficiency-report`
- `list-sprints`
- `get-sprint`
- `get-sprint-report`
- `get-sprint-health-report`
- `get-active-sprint`
- `search-sprint`
- `add-comment`
- `get-comments`
- `add-worklog`
- `get-transitions`
- `transition-issue`
- `list-statuses`
- `get-issue-history`
- `get-related-issues`
- `link-issues`
- `get-version`
- `list-project-versions`
- `get-development-info`
- `download-attachment`

## Gemini agent with ADK Go

The Gemini integration does not re-implement Jira logic. It connects Gemini to the local MCP server through ADK's `mcptoolset`, so MCP remains the single tool authority.

The agent now keeps persistent sessions in a local SQLite database. In console mode, the active conversation is keyed by `--user-id` plus either an explicit `--session-id` or `--resume-last`. This lets you continue the same chat across process restarts instead of losing context when the binary exits.

The command now uses explicit subcommands: use `console` for terminal chat, `web` for the ADK web launcher, and `sessions` for session management.

Run the agent:

```bash
jiraforge-agent console --env .env
```

Resume the latest session for a user:

```bash
jiraforge-agent console --env .env --user-id alice --resume-last
```

Create or reopen a named session:

```bash
jiraforge-agent console --env .env --user-id alice --session-id release-audit
```

Override the model:

```bash
jiraforge-agent console --env .env --model gemini-2.5-flash
```

You can override that if you want to point ADK to another command:

```bash
jiraforge-agent console --env .env --mcp-command jiraforge --mcp-args "--http_port 3000"
```

Change the session database location if needed:

```bash
jiraforge-agent console --env .env --session-db ./var/jiraforge-agent.db
```

By default the session database lives in the user's config directory, for example under `~/Library/Application Support/jiraforge/` on macOS.

In addition to exact session resume, the agent can also recall relevant snippets from older sessions for the same `--user-id` when they help answer the current request.

Run the web launcher with additional ADK launcher arguments after `--`:

```bash
jiraforge-agent web --env .env -- --host 127.0.0.1 --port 8080
```

You can manage the local session store directly without loading Jira or Gemini configuration:

```bash
jiraforge-agent --session-db ./var/jiraforge-agent.db sessions list --user-id alice
jiraforge-agent --session-db ./var/jiraforge-agent.db sessions inspect --user-id alice --session-id release-audit
jiraforge-agent --session-db ./var/jiraforge-agent.db sessions delete --user-id alice --session-id release-audit
```

Mutating Jira tools require confirmation in the ADK toolset layer.

## Development

Run tests:

```bash
go test ./...
```

Capabilities currently exposed through the shared Jira service (and therefore through MCP, CLI, and the Gemini agent):

- issue CRUD (create, child issue, update, delete) and issue type listing
- search and JQL
- sprints: list, get, active sprint, search by name, sprint report
- management reports: aging, blocked issues, flow efficiency, sprint health, sprint workload per assignee, utilization (worklogs), cycle time / throughput, team WIP snapshot
- comments and worklogs (including worklog audit per issue)
- workflow: transitions, transition execution, status catalog, issue history (changelog)
- issue relationships: related issues and link creation
- versions: get and list project versions
- development information: branches, pull requests, and commits linked to an issue
- attachment download to local temp storage
- MCP prompts: `issue_development_tree`, `release_development_overview`

For end-to-end Tech Lead workflows (daily WIP, sprint workload review, utilization,
cycle time) see [docs/team_lead_workflow.md](docs/team_lead_workflow.md).

## License

MIT. See `LICENSE`.
