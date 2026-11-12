# mia-brains(1)

## NAME

mia-brains - the registered brains

## DESCRIPTION

Lists the brains registered on this machine, sorted by name, with whether each
folder exists and whether it has its own config, `.mia_dna.json`.

The registry is `$XDG_DATA_HOME/memoria/brains.json`, by default
`~/.local/share/memoria/brains.json`. It holds absolute paths and belongs to
one machine; it is written only by mia-register-brain(1) and
mia-deregister-brain(1).

## OUTPUT

```
{ "ok": true, "result": [ { "name": "work", "location": "/home/me/notes/work", "exists": true, "has_dna": false } ] }
```

An empty registry answers an empty list.

## ERRORS

A registry that cannot be read or is not valid JSON is an error naming the
file.

## EXIT STATUS

`0` when the registry was read, `1` otherwise.

## SEE ALSO

mia-register-brain(1), mia-deregister-brain(1)
