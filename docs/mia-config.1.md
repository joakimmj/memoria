# mia-config(1)

## NAME

mia-config - the merged config, or the built-in defaults alone

## DESCRIPTION

Answers the configuration every other command uses: the built-in defaults,
then the user's config, merged as described in mia-config(5). With
`--defaults` it answers the built-in defaults alone, to copy the keys worth
changing from.

Nothing writes a config file: a missing one means the defaults.

## OUTPUT

```
{ "ok": true, "result": { "engrams": { "date_format": "YYYY-MM-DD", … } } }
```

The result has the shape of the config files, with every key present.

## ERRORS

A config file that cannot be read, is not valid JSON, holds an unknown key or
a value of the wrong type is an error naming the file.

## EXIT STATUS

`0` when the config was read, `1` otherwise.

## EXAMPLES

```
mia config
mia config --defaults
```

## SEE ALSO

mia-config(5)
