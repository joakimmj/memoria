# mia-engram(1)

## NAME

mia-engram - one engram: its atlas entry, backlinks, content and date formats

## DESCRIPTION

Answers one engram, read as described under "Engram arguments" in mia(1):
its entry in the brain's atlas, brought up to date first; the engrams
linking to it; its content; and the brain's two date formats, for reading
the engram's date — `filename_date_format` for a dated filename's prefix,
`date_format` for dates written in the engram.

`--body` answers the content without the frontmatter, the `SYNAPSES` block
and the blank line after them.

## OUTPUT

```
{ "ok": true, "result": {
    "brain": "work", "file": "2026-08-01_project_x.md",
    "path": "/home/me/notes/work/2026-08-01_project_x.md",
    "entry": { "title": "Project X", "synapses": { "down": [], "up": [] }, "links": [], "modified": "2026-08-01T10:00:00Z", "hash": "a1b2c3d4e5f60718" },
    "backlinks": [ "2026-07-31_meeting_notes.md" ],
    "content": "…",
    "date_format": "YYYY-MM-DD", "filename_date_format": "YYYY-MM-DD" } }
```

## ERRORS

- `not_found`: no such engram, or its file could not be read into the atlas.
- `ambiguous`, `not_in_brain`, `not_an_engram`: see "Engram arguments" in
  mia(1).

## EXIT STATUS

`0` when the engram was answered, `1` otherwise.

## EXAMPLES

```
mia engram project_x
mia engram ./2026-08-01_project_x.md --body
```

## SEE ALSO

mia-engrams(1), mia-attach-synapse(1)
