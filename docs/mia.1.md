# mia(1)

## NAME

mia - the memoria command line

## DESCRIPTION

**mia** reads and writes brains: folders of markdown notes called engrams,
connected by synapses and concepts. Every command is one run: it starts from
its arguments, the working directory and the files, and keeps nothing after it
exits.

Run `mia help <command>` for a command's page, or `mia <command> --help`,
which is the same page. `--help` loads no configuration, so it works when
nothing else does.

## OUTPUT

Every command except **help** writes exactly one JSON object to stdout:

```
{ "ok": true, "result": … }
{ "ok": false, "error": "…", "code": "…" }
```

`code` is present wherever a caller can act on the failure. Diagnostics go to
stderr, so stdout is always one object.

## ERRORS

- `usage`: the command line is wrong: an unknown command, a missing or extra
  argument, an unknown flag.
- `not_found`: a named thing does not exist.

## EXIT STATUS

`0` when the result is ok, `1` otherwise.

## SEE ALSO

mia-help(1), mia-commands(1)
