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

**Engram arguments.** An `<engram>` argument is read one way by every
command:

- holding a `/`, such as `./x.md` or `~/notes/work/x.md`: a path to the file.
  Its folder decides the brain; with brains nested in one another, the one
  the file sits directly in. The file must be a `.md` file directly in a
  brain, not in a subfolder.
- otherwise, such as `x` or `x.md`: an engram's name, `.md` optional, looked
  up in `--brain` when given, else in every brain.

A command taking two engrams finds them in one brain: `--brain` when given;
else a path's brain, the other looked up there; else the one brain holding
both names.

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
- `brain`: a command that writes in one brain was given no `--brain`.
- `not_found`: a named thing does not exist: a brain, an engram, a field, a
  help topic.
- `collision`: the thing to add already exists, such as a brain's name or
  folder.
- `ambiguous`: an engram name found in more than one brain; the error lists
  them.
- `cross_brain`: two engrams that must share a brain do not.
- `not_in_brain`: a path outside every brain, or outside the one `--brain`
  names.
- `not_an_engram`: a path in a brain that is not a `.md` file directly in it.
- `problems`: `check --strict` found something wrong; the problems are the
  result.

## EXIT STATUS

`0` when the result is ok, `1` otherwise.

## SEE ALSO

mia-help(1), mia-commands(1), mia-brains(1), mia-engrams(1), mia-engram(1),
mia-tasks(1), mia-check(1), mia-rebuild(1), mia-attach-synapse(1),
mia-config(5)
