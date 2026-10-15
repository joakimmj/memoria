# mia-man(1)

## NAME

mia-man - the manual pages written as roff

## DESCRIPTION

Writes every manual page as roff into the directory given, one subdirectory
per section: `man1/mia.1`, `man1/mia-help.1`, and so on. Existing files of the
same name are replaced. Once the directory is on `MANPATH`, `man mia` works.

## OUTPUT

```
{ "ok": true, "result": { "dir": "/abs/dir", "pages": ["man1/mia.1", …] } }
```

`pages` are relative to `dir`.

## ERRORS

- `usage`: `--dir` was not given.

## EXIT STATUS

`0` when every page was written, `1` otherwise.

## EXAMPLES

```
mia man --dir ~/.local/share/man
```

## SEE ALSO

mia-help(1)
