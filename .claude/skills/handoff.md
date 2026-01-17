# /handoff

Save state for continuation in new chat (use when context ~10-15% remaining).

## Description

Create context handoff for session continuation.

## Allowed Tools

- Read
- Write
- Glob

## Execute

1. Summarize current project/phase
2. Note key files and decisions
3. Save to `.claude/handoffs/handoff-[date]-[time].md`
4. Provide continuation prompt

## Handoff Format

```markdown
# Context Handoff - [Date]

## Current Project / Plan

[Brief description of the project and current plan]

## Current Phase

[What phase of work we're in]

## Work Completed This Session

- [Item 1]
- [Item 2]
- [Item 3]

## Key Files

- `path/to/file1` - [description]
- `path/to/file2` - [description]

## Decisions Made

- [Decision 1]
- [Decision 2]

## Next Steps

1. [Step 1]
2. [Step 2]
3. [Step 3]

## Continuation Prompt

[Pre-written prompt to paste into new session]
```

## Output

```
Handoff saved to: .claude/handoffs/handoff-YYYY-MM-DD-HHMM.md

To continue, paste:
---
Resume from handoff: [path]
Context: [brief]
Next: [action]
---
```
