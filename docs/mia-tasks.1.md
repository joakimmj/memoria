# mia-tasks(1)

## NAME

mia-tasks - every task, filtered by state or engram

## DESCRIPTION

Lists every task of every brain, or of the brains `--brain` names, in one
list. A task is a list item whose marker is one of `engrams.task_markers`
(mia-config(5)), such as `- [ ] Call Alice`, outside fenced code and the
`SYNAPSES` block. Each brain's atlas is brought up to date first; a brain
that cannot be read is skipped with a warning, unless `--brain` names it.

Filters combine; one left out filters nothing:

- `--state`: `not_done` or `done`.
- `--engram`: one engram's tasks, read as described under "Engram arguments"
  in mia(1). It decides the brain; with `--brain`, the engram must be in it.

Ordered by brain, then by engram, the most recently changed first, then by
line.

## OUTPUT

```
{ "ok": true, "result": { "tasks": [ { "brain": "work", "engram": "2026-10-20_weekly_sync.md", "line": 14, "text": "Send budget draft", "state": "not_done" } ], "done": 0, "total": 1 } }
```

`line` counts from 1; `text` is the task without its marker. `done` and
`total` count the tasks answered.

## ERRORS

- `usage`: `--state` is neither `not_done` nor `done`, or `--engram` comes
  with more than one `--brain`.
- `not_found`: a brain `--brain` names is not registered, or no such engram.
- `ambiguous`, `not_in_brain`, `not_an_engram`: see "Engram arguments" in
  mia(1).

## EXIT STATUS

`0` when the tasks were listed, `1` otherwise.

## EXAMPLES

```
mia tasks --state not_done
mia tasks --engram 2026-10-20_weekly_sync
```

## SEE ALSO

mia-engrams(1), mia-config(5)
