# mia-engrams(1)

## NAME

mia-engrams - every engram, newest first

## DESCRIPTION

Lists every engram of every brain, or of the brains `--brain` names; repeat
it for several. Each brain's atlas is brought up to date first. A brain that
cannot be read is skipped with a warning on stderr, unless `--brain` names
it, which makes it an error.

Newest first by the time each file last changed, then by brain and filename.

## OUTPUT

```
{ "ok": true, "result": [ { "brain": "work", "file": "2026-08-01_project_x.md", "title": "Project X", "modified": "2026-08-01T10:00:00Z" } ] }
```

`title` is the engram's first heading, or its filename without `.md`.
No engrams is an empty list.

## ERRORS

- `not_found`: a brain `--brain` names is not registered.

## EXIT STATUS

`0` when the engrams were listed, `1` otherwise.

## EXAMPLES

```
mia engrams
mia engrams --brain work --brain personal
```

## SEE ALSO

mia-engram(1), mia-tasks(1)
