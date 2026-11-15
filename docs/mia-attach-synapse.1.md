# mia-attach-synapse(1)

## NAME

mia-attach-synapse - link an engram to another through a synapse field, and back through its inverse

## DESCRIPTION

Adds `<target>` to the synapse field `<field>` of `<source>`, in the
`SYNAPSES` block at the top of the engram. When the field has an `inverse`
(mia-config(5)), `<source>` is added to that field of `<target>` too, so
`mia attach-synapse a down b` also gives `b` the `up` link back to `a`.

A field with `list` set to `false` holds one engram: attaching replaces the
old one, which loses its link back. A link already there is left alone, so
attaching twice changes nothing, and attaching a link that has no way back
adds only the missing half.

Both engrams are in one brain; see "Engram arguments" in mia(1). Every file
is read before any is written. The block is rewritten with its fields sorted;
a field the config does not name is kept, while a line that is not a field,
and a link in a field that is not to an engram in the brain, are dropped.

## OUTPUT

```
{ "ok": true, "result": { "brain": "work", "source": "/home/me/notes/work/a.md", "field": "down", "target": "/home/me/notes/work/b.md" } }
```

## ERRORS

- `not_found`: no such engram, or `<field>` is not a synapse field of the
  brain's config.
- `usage`: `<source>` and `<target>` are the same engram.
- `ambiguous`, `cross_brain`, `not_in_brain`, `not_an_engram`: see "Engram
  arguments" in mia(1).

## EXIT STATUS

`0` when both engrams are linked, `1` otherwise.

## EXAMPLES

```
mia attach-synapse project_x down meeting_notes
mia attach-synapse --brain work ./2026-08-01_project_x.md down meeting_notes
```

## SEE ALSO

mia(1), mia-config(5), mia-create-engram(1)
