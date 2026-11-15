# mia-check(1)

## NAME

mia-check - rebuild each atlas and answer what is wrong, writing no engram

## DESCRIPTION

Rebuilds the atlas of every brain, or of the brains `--brain` names, parsing
every engram again, and answers the problems found. No engram is written;
mia-rebuild(1) with `--fix` repairs what can be repaired. A brain that cannot
be read is skipped with a warning, unless `--brain` names it.

The problems:

- `broken_synapse`: a synapse to an engram that does not exist, on the
  field's line.
- `missing_inverse`: a synapse whose inverse field (mia-config(5)) on the
  other engram does not link back, on the field's line.
- `broken_link`: a link in the body to an engram that does not exist, on the
  first line linking it.

Ordered by brain, engram, line, then kind.

## OUTPUT

```
{ "ok": true, "result": [ { "brain": "work", "kind": "broken_link", "engram": "a.md", "file": "/home/me/notes/work/a.md", "line": 12, "text": "broken link: x.md" } ] }
```

`line` counts from 1. No problems is an empty list.

## ERRORS

- `problems`: with `--strict`, there is at least one problem. The problems
  are the `result`, beside the error.
- `not_found`: a brain `--brain` names is not registered.

## EXIT STATUS

`0` when the brains were checked, and with `--strict` nothing was wrong; `1`
otherwise.

## EXAMPLES

```
mia check
mia check --brain work --strict
```

## SEE ALSO

mia-rebuild(1), mia-attach-synapse(1)
