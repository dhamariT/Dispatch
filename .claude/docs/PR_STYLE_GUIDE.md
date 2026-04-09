# Pull Request Description Style Guide

This guide documents the PR description style used in the Dispatch repository.

## PR Title Format

Format: `type(scope): description`.

- Types: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `build`, `ci`, `chore`, `revert`
- Scopes must be a real path (directory or file stem) containing all changed files
- Omit scope if changes span multiple top-level directories

Examples:

- `feat: add snapshot diff engine`
- `fix(fleet): handle offline devices in Balena polling`
- `perf(snapshot): add index on snapshots.device_id`
- `docs: update architecture doc with agent protocol`
- `refactor(api): extract metric validation middleware`

## PR Description Structure

### Default Pattern: Keep It Concise

Most PRs use a simple 1-2 paragraph format:

```markdown
[Brief statement of what changed]

[One sentence explaining technical details or context if needed]
```

**Example (bugfix):**

```markdown
Previously, when a device went offline mid-deploy, the snapshot engine
would hang waiting for the after-snapshot indefinitely.

Add a configurable timeout to snapshot collection that marks offline
devices as "incomplete" and continues the rollout evaluation.
```

**Example (dependency update):**

```markdown
Changes from https://github.com/upstream/repo/pull/XXX/
```

**Example (docs correction):**

```markdown
Adds the initial Balena API client for polling device vitals.
Fetches CPU, memory, temperature, and storage on a configurable interval.
```

### For Complex Changes: Use "Summary", "Problem", "Fix"

Only use structured sections when the change requires significant explanation:

```markdown
## Summary
Brief overview of the change

## Problem
Detailed explanation of the issue being addressed

## Fix
How the solution works
```

**Example (API documentation fix):**

```markdown
## Summary
Refactor snapshot storage to use TimescaleDB hypertables for time-series queries...

## Problem
Querying snapshot history for a single device over 30 days was slow with plain Postgres...

## Fix
Convert the snapshots table to a TimescaleDB hypertable partitioned by timestamp...
```

### For Large Refactors: Lead with Context

When rewriting significant documentation or code, start with the problems being fixed:

```markdown
This PR rewrites [component] for [reason].

The previous [component] had [specific issues]: [details].

[What changed]: [specific improvements made].

[Additional changes]: [context].

Refs #[issue-number]
```

**Example (major documentation rewrite):**

- Started with "This PR rewrites the rollout engine to support multi-wave deploys"
- Listed specific limitations of the previous single-canary approach
- Explained the new wave promotion logic
- Referenced related issue

## What to Include

### Always Include

1. **Link Related Work**
   - `Closes #XXX`
   - `Depends on #XXX`
   - `Refs #XXX` (for general reference)

2. **Performance Context** (when relevant)

   ```markdown
   Balena API polling with 50 devices was taking ~8s per cycle.
   Batching requests brings it down to ~1.2s.
   ```

3. **Migration Warnings** (when relevant)

   ```markdown
   **NOTE**: This migration adds a TimescaleDB hypertable on `snapshots`.
   For existing deployments with historical data, this may take a moment.
   ```

4. **Visual Evidence** (for UI changes)

   ```markdown
   <img width="1281" height="425" alt="image" src="..." />
   ```

### Never Include

- ❌ **Test plans** - Testing is handled through code review and CI
- ❌ **"Benefits" sections** - Benefits should be clear from the description
- ❌ **Implementation details** - Keep it high-level
- ❌ **Marketing language** - Stay technical and factual
- ❌ **Bullet lists of features** (unless it's a large refactor that needs enumeration)

## Special Patterns

### Simple Chore PRs

For straightforward updates (dependency bumps, minor fixes):

```markdown
Changes from [link to upstream PR/issue]
```

Or:

```markdown
Reference:
[link explaining why this change is needed]
```

### Bug Fixes

Start with the problem, then explain the fix:

```markdown
[What was broken and why it matters]

[What you changed to fix it]
```

### Dependency Updates

Dependabot PRs are auto-generated - don't try to match their verbose style for manual updates. Instead use:

```markdown
Changes from https://github.com/upstream/repo/pull/XXX/
```

## Creating PRs as Draft

**IMPORTANT**: Unless explicitly told otherwise, always create PRs as drafts using the `--draft` flag:

```bash
gh pr create --draft --title "..." --body "..."
```

After creating the PR, encourage the user to review it before marking as ready:

```text
I've created draft PR #XXXX. Please review the changes and mark it as ready for review when you're satisfied.
```

This allows the user to:

- Review the code changes before requesting reviews from maintainers
- Make additional adjustments if needed
- Ensure CI passes before notifying reviewers
- Control when the PR enters the review queue

Only create non-draft PRs when the user explicitly requests it or when following up on an existing draft.

## Key Principles

1. **Always create draft PRs** - Unless explicitly told otherwise
2. **Be concise** - Default to 1-2 paragraphs unless complexity demands more
3. **Be technical** - Explain what and why, not detailed how
4. **Link everything** - Issues, PRs, upstream changes, Notion docs
5. **Show impact** - Metrics for performance, screenshots for UI, warnings for migrations
6. **No test plans** - Code review and CI handle testing
7. **No benefits sections** - Benefits should be obvious from the technical description

## Examples by Category

### Performance Improvements

Includes query timing metrics and explains the index solution

### Bug Fixes

Describes broken behavior then the fix in two sentences

### Documentation

- **Major rewrite**: Long form explaining inaccuracies and improvements
- **Simple correction**: One sentence for simple correction

### Features

Simple statement of what was added and dependencies

### Refactoring

Explains why client-side sorting is now redundant

### Configuration

Adds guidelines with issue reference
