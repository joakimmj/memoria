# mia-config(5)

## NAME

mia-config - config.json and .mia_dna.json, the memoria configuration files

## DESCRIPTION

Configuration comes in three tiers, merged in order, later over earlier:

- the built-in defaults, shown by `mia config --defaults`
- the user's config, `$XDG_CONFIG_HOME/memoria/config.json`, by default
  `~/.config/memoria/config.json`, the same on Linux and macOS
- a brain's config, `.mia_dna.json` at the brain's root

Both files are JSON objects of the same shape, so a key means the same in
either. A file holds only what differs from the tier below it; nothing writes
or seeds one. A missing file is skipped. A file that cannot be read, is not
valid JSON, holds an unknown key or a value of the wrong type is an error
naming the file. Keys are snake_case.

Merging:

- Objects merge by key; lists and other values replace what is below
  wholesale.
- A key set to `null` is not the same as a key left out. Left out, it follows
  the tier below; set to `null`, it takes the built-in default, skipping the
  user's config.

Date formats use the tokens `YYYY`, `YY`, `MM`, `DD`, `HH`, `mm` and `ss`;
anything else is written literally.

The keys:

- `engrams.date_format`: the format of the `%date%` placeholder. Default
  `YYYY-MM-DD`.
- `engrams.filename.prefix`: what starts a new engram's filename, joined to
  the slug of its title with `_`: `date` or `none`. Default `date`.
- `engrams.filename.date_format`: the format of the `date` prefix. Default
  `YYYY-MM-DD`.
- `engrams.content_template`: the text a new engram starts with.
  `%title%` is the title as given, `%date%` today, and `%body%`
  marks where the body goes and is removed. Default `# %title%`, a blank
  line, then `%body%`.

## EXAMPLES

A brain without date prefixes:

```
{ "engrams": { "filename": { "prefix": "none" } } }
```

A user who writes dates day first, and a brain that goes back to the
default:

```
{ "engrams": { "date_format": "DD.MM.YYYY" } }
{ "engrams": { "date_format": null } }
```

## SEE ALSO

mia-config(1)
