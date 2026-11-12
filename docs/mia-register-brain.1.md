# mia-register-brain(1)

## NAME

mia-register-brain - register a folder as a brain, creating it when missing

## DESCRIPTION

Adds the folder at `<path>` to the registry under `--name`, or under the
folder's basename. A relative path is taken from the working directory and
stored absolute, so the brain resolves the same from anywhere. A missing
folder is created; an existing one is left as it is.

A name is not empty, holds no `/` and does not start with `-`.

## OUTPUT

```
{ "ok": true, "result": { "name": "work", "location": "/home/me/notes/work" } }
```

## ERRORS

- `collision`: the name, or the folder, is already registered.
- `usage`: the name is not valid.

## EXIT STATUS

`0` when the brain was registered, `1` otherwise.

## EXAMPLES

```
mia register-brain ~/notes/work
mia register-brain . --name personal
```

## SEE ALSO

mia-brains(1), mia-deregister-brain(1)
