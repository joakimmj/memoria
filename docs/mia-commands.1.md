# mia-commands(1)

## NAME

mia-commands - every command with its arguments, as a schema

## DESCRIPTION

Answers the command table: the same list, in the same order, that dispatch,
argument parsing and these manual pages are built from. The order is stable.

## OUTPUT

```
{ "ok": true, "result": [
  { "name": "man", "summary": "…",
    "flags": [{ "name": "dir", "value": "dir", "summary": "…", "required": true }] },
  …
] }
```

Per command: `name`, `summary`, `args` (positional, in order: `name`,
`summary`, `optional`), `flags` (`name`, `value` for a flag that takes one,
`summary`, `required`), and `text` for a command that prints text rather than
JSON. Fields that are empty or false are left out.

## EXIT STATUS

`0`.

## SEE ALSO

mia(1)
