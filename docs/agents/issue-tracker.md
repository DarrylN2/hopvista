# Issue tracker: GitHub

Specs, decisions, and implementation tickets for this repo live as GitHub issues.
Use the repository's `origin` remote and the `gh` CLI when available.
If `gh` is unavailable, use the connected GitHub app's issue tools.

## Conventions

- Create an issue: `gh issue create --title "..." --body-file <path>`.
- Read an issue: `gh issue view <number> --comments`.
- List issues: `gh issue list --state open --json number,title,body,labels,assignees`.
- Comment: `gh issue comment <number> --body-file <path>`.
- Add or remove labels: `gh issue edit <number> --add-label "..."` or `--remove-label "..."`.
- Close: `gh issue close <number> --comment "..."`.
- Use a temporary UTF-8 file with `--body-file` for multiline issue text.
- With the connected GitHub app, pass multiline bodies directly to its issue tools.

## Pull requests as a triage surface

PRs as a request surface: no.

## Skill routing

- "Publish to the issue tracker" means create a GitHub issue.
- "Fetch the relevant ticket" means read its GitHub issue and comments.

## Wayfinding operations

- Map: one issue labelled `wayfinder:map`.
- Child ticket: a GitHub sub-issue of the map, labelled `wayfinder:research`,
  `wayfinder:prototype`, `wayfinder:grilling`, or `wayfinder:task`.
  If sub-issues are unavailable, use a task list in the map and
  `Part of #<map>` in each ticket.
- Blocking: use GitHub issue dependencies. For the API, fetch the blocker's
  numeric database `id`, then add it to the child's `blocked_by` dependencies.
  If dependencies are unavailable, put `Blocked by: #<n>` in the ticket body.
- Frontier: the map's open, unassigned child tickets with no open blockers,
  in map order.
- Claim: assign the ticket to the developer before working it.
- Resolve: post the answer as a comment, close the ticket, then append a
  one-line gist and link to the map's Decisions-so-far section.
