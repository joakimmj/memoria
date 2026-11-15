# mia-rebuild(1)

## NAME

mia-rebuild - rebuild the atlas, repairing what can be repaired first with --fix

## DESCRIPTION

Rebuilds the atlas of the brain `--brain` names, parsing every engram again as
if no atlas existed, and answers the problems left, as mia-check(1) does.

With `--fix`, it repairs first:

1. writes every missing inverse, as mia-attach-synapse(1) would;
2. regenerates every existing `SYNAPSES` block from the config, which adds a
   line for a synapse field configured since the engram was written.

Engrams without a block are left alone. A broken synapse or link has no
repair, since removing it would lose what it says; it stays in the problems.
A repair that fails is a warning on stderr, and its problem stays too.

## OUTPUT

```
{ "ok": true, "result": [ { "brain": "work", "kind": "broken_synapse", "engram": "a.md", "file": "/home/me/notes/work/a.md", "line": 2, "text": "broken synapse: down → x.md" } ] }
```

## ERRORS

- `brain`: no `--brain` given; the error lists the registered brains.
- `not_found`: the brain is not registered.

## EXIT STATUS

`0` when the atlas was rebuilt, `1` otherwise.

## EXAMPLES

```
mia rebuild --brain work
mia rebuild --brain work --fix
```

## SEE ALSO

mia-check(1), mia-config(5)
