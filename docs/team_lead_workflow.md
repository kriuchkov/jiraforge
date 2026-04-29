# Team Lead Workflow with JiraForge

This guide explains how a Tech Lead / Team Lead can use **JiraForge** — through the
**MCP server** (`jiraforge`), the **CLI** (`jiraforge-cli`), or the natural-language
**agent** (`jiraforge-agent`) — to run the most common team-lead workflows without
opening Jira UI.

The goal: fast, reproducible, JSON-or-text answers for daily stand-ups, sprint reviews,
and capacity planning.

---

## 1. Capabilities at a glance

| Workflow | MCP tool | CLI command |
|---|---|---|
| Daily WIP snapshot | `jira_get_team_wip_snapshot` | `get-team-wip-snapshot` |
| Sprint workload per assignee | `jira_get_sprint_workload_report` | `get-sprint-workload-report` |
| Sprint health | `jira_get_sprint_health_report` | `get-sprint-health-report` |
| Utilization (worklogs) | `jira_get_utilization_report` | `get-utilization-report` |
| Cycle time / throughput | `jira_get_cycle_time_report` | `get-cycle-time-report` |
| Aging issues | `jira_get_aging_report` | `get-aging-report` |
| Blocked issues | `jira_get_blocked_issues_report` | `get-blocked-issues-report` |
| Flow efficiency | `jira_get_flow_efficiency_report` | `get-flow-efficiency-report` |
| Worklog audit | `jira_get_worklogs` | `get-worklogs` |

All CLI commands accept `--output text|json` (default `text`). MCP tools always return text.

---

## 2. Daily stand-up: WIP snapshot

**Question:** *Who is working on what right now? Is anyone over their WIP limit?*

### MCP

```json
{
  "tool": "jira_get_team_wip_snapshot",
  "input": {
    "project_key": "PROJ",
    "wip_limit": 3
  }
}
```

### CLI

```bash
jiraforge-cli get-team-wip-snapshot --project-key PROJ --wip-limit 3
```

### Agent prompt

> "Show me the current WIP snapshot for project PROJ. Flag anyone with more than 3 issues in flight."

The snapshot groups in-progress issues per assignee, breaks them down by status,
and marks anyone above `--wip-limit` with a warning.

---

## 3. Sprint review: workload per assignee

**Question:** *In sprint 42, who committed how much, what's done, what slipped, and what was added mid-sprint?*

### MCP

```json
{ "tool": "jira_get_sprint_workload_report", "input": { "sprint_id": "42" } }
```

### CLI

```bash
jiraforge-cli get-sprint-workload-report --sprint-id 42
```

### Agent prompt

> "Build the sprint workload report for sprint 42 and tell me who is at risk of not finishing."

Output includes per-assignee committed/completed/incomplete counts, story-point
estimates, commitment-completion ratio, status breakdown, and the keys of issues
added during the sprint.

---

## 4. Capacity & utilization (worklogs)

**Question:** *How many hours did the team log this week, and what did each person spend time on?*

### MCP

```json
{
  "tool": "jira_get_utilization_report",
  "input": {
    "project_key": "PROJ",
    "window_days": 7,
    "top_issues_per_user": 5
  }
}
```

### CLI

```bash
jiraforge-cli get-utilization-report --project-key PROJ --window-days 7
jiraforge-cli get-utilization-report --jql "project = PROJ AND worklogAuthor = alice" \
  --start-date 2026-04-01 --end-date 2026-04-30
```

### Agent prompt

> "Give me the team utilization for the last 7 days in project PROJ, with the top 5 issues per person."

The report iterates worklogs in the window, aggregates hours per author, and
returns daily-average hours plus the top issues each person logged time on.

> ⚠ Worklogs are scanned per matching issue. Use `--max-results` to bound the
> issue scan; for a focused query, prefer a tight `--jql` filter.

---

## 5. Cycle time & throughput

**Question:** *How long does work take from "In Progress" to "Done"? What's our weekly throughput?*

### MCP

```json
{
  "tool": "jira_get_cycle_time_report",
  "input": {
    "project_key": "PROJ",
    "start_statuses": "In Progress",
    "done_statuses": "Done, Closed, Resolved",
    "window_days": 30
  }
}
```

### CLI

```bash
jiraforge-cli get-cycle-time-report --project-key PROJ --window-days 30
jiraforge-cli get-cycle-time-report --jql "project = PROJ AND issuetype = Bug" \
  --start-status "In Progress" --done-status Done --done-status Closed
```

### Agent prompt

> "Compute cycle time and throughput for project PROJ over the last 30 days, broken down by issue type."

Summary returns average / median / p85 cycle hours, average lead hours,
total throughput, and throughput per week. Per-assignee and per-type
breakdowns highlight where work flows fastest or stalls.

---

## 6. Worklog audit

**Question:** *What worklogs were filed against PROJ-123 since April 1?*

### MCP

```json
{
  "tool": "jira_get_worklogs",
  "input": { "issue_key": "PROJ-123", "started_after": "2026-04-01" }
}
```

### CLI

```bash
jiraforge-cli get-worklogs --issue-key PROJ-123 --started-after 2026-04-01
```

---

## 7. Putting it all together: a Tech Lead's morning routine

Run these in order from your terminal (or chat with `jiraforge-agent`):

```bash
# 1. Fresh stand-up snapshot
jiraforge-cli get-team-wip-snapshot --project-key PROJ --wip-limit 3

# 2. Anything stale we should escalate?
jiraforge-cli get-aging-report --project-key PROJ --min-days-in-status 5

# 3. Anything blocked?
jiraforge-cli get-blocked-issues-report --project-key PROJ

# 4. Where is flow stalling?
jiraforge-cli get-flow-efficiency-report --project-key PROJ --window-days 14
```

Or, with the agent:

> "Run my morning team-lead briefing for PROJ:
> 1. Current WIP per person, flag overload above 3.
> 2. Issues stuck >5 days in any status.
> 3. Currently blocked issues.
> 4. Last 14 days flow efficiency.
> Summarize the top 5 risks."

---

## 8. Configuring MCP for your IDE

The MCP server is launched via `jiraforge` and reads Atlassian credentials from
the environment (or an `.env` file). See [README.md](../README.md) for full
configuration. Once registered, every tool above becomes available to the
LLM running inside your IDE / `jiraforge-agent`.

A minimal `.mcp.json` snippet:

```json
{
  "servers": {
    "jiraforge": {
      "command": "jiraforge",
      "env": { "JIRA_BASE_URL": "...", "JIRA_EMAIL": "...", "JIRA_API_TOKEN": "..." }
    }
  }
}
```

---

## 9. Tips

- **JQL beats project_key.** When you can express the slice you care about as
  JQL, use `--jql` / `"jql"` to bypass the auto-generated query.
- **Default windows are short.** Utilization defaults to 7 days, cycle time to
  30 days. Override with `--window-days` or explicit `--start-date` / `--end-date`.
- **Use JSON for piping.** `jiraforge-cli ... --output json | jq` keeps these
  reports composable with `jq`, spreadsheets, or follow-up tooling.
- **Status names matter.** Cycle time honours your real workflow status
  names — pass them via `--start-status` / `--done-status` to match how the
  team actually transitions issues.
