# mia-create-engram(1)

## NAME

mia-create-engram - a new engram from the content template

## DESCRIPTION

Writes a new engram into the brain named by `--brain`. Its filename is the
`engrams.filename.prefix` of mia-config(5) joined to the slug of `--title`
with `_`:

- `date`: the date in `engrams.filename.date_format`, as in
  `2026-08-01_project_x.md`. Without a title, the date alone.
- `none`: the slug alone, as in `project_x.md`. A title is required.

The slug is the title trimmed and lowercased, with Latin letters written in
ASCII and every run of anything else turned into one `_`. The filename is the
engram's identity: an existing file is never overwritten.

The file holds `engrams.content_template` with its placeholders expanded:
`%title%` is the title as given, `%date%` today in `engrams.date_format`, and
`%body%` marks where `--body` goes. Without `%body%`, the body goes at the
end. `--body -` reads the body from stdin.

## OUTPUT

```
{ "ok": true, "result": { "brain": "work", "path": "/home/me/notes/work/2026-08-01_project_x.md", "body": { "line": 3, "col": 0 }, "created": true } }
```

`body` is where the body starts: `line` counts from 1, `col` counts bytes from
0, so an editor can open the file with the cursor there.

## ERRORS

- `brain`: no `--brain` given; the error lists the registered brains.
- `not_found`: the brain is not registered.
- `collision`: the file already exists.
- `empty_slug`: the title has no letter or digit.
- `prefix`: the brain's filename prefix is not supported.
- `template`: the content template is not valid: it holds more than one
  `%body%`.
- `usage`: no title, where the prefix needs one.

## EXIT STATUS

`0` when the engram was written, `1` otherwise.

## EXAMPLES

```
mia create-engram --brain work --title "Project X"
echo "first thoughts" | mia create-engram --brain work --title "Project X" --body -
mia create-engram --brain journal
```

## SEE ALSO

mia-config(5), mia-brains(1)
