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
valid JSON, holds an unknown key, a value of the wrong type or a synapse name
other than letters, digits and `_` is an error naming the file. Keys are snake_case.

Merging:

- Objects merge by key; lists and other values replace what is below
  wholesale.
- A key set to `null` is not the same as a key left out. Left out, it follows
  the tier below; set to `null`, it takes the built-in default, skipping the
  user's config.
- An entry of `synapses` set to `null` is removed and stays removed: the one
  way to drop a default field. A later tier may add it again.

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
- `synapses.<name>`: a synapse field, a link from an engram to others,
  written as `<name>` in the engram's `SYNAPSES` block. A name is letters,
  digits and `_`. Default `down` and `up`, each the other's inverse.
- `synapses.<name>.inverse`: the field kept in sync on the target, so
  attaching `b` to `a`'s `down` adds `a` to `b`'s `up`. Default none.
- `synapses.<name>.list`: whether the field holds several engrams. Default
  `true`.
- `synapses.<name>.show_empty`: whether the field's line is written with no
  engrams in it. Default `true`.

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

A brain that drops `up` and adds `related`, a field with no inverse that is
only written when it holds something:

```
{ "synapses": { "up": null, "related": { "show_empty": false } } }
```

## SEE ALSO

mia-config(1)
