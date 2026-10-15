# mia-version(1)

## NAME

mia-version - the version and commit mia was built from

## DESCRIPTION

Reads the build information the Go toolchain stamps into the binary. A binary
installed with `go install …@vX.Y.Z` reports that version; one built in a git
checkout reports a version derived from the nearest tag, and the commit.
Anything else reports `dev`.

## OUTPUT

```
{ "ok": true, "result": { "version": "v0.1.0", "commit": "…" } }
```

`commit` is the full commit hash, or empty when the build has none.

## EXIT STATUS

`0`.

## SEE ALSO

mia(1)
