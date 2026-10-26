# Go notes

Go concepts learned while building memoria, one line each, by step.

## Phase 0 — the project

- **Modules and packages:** a module (`go.mod`) is the unit of versioning, like a Maven artifact; a package is one directory, and every file in it shares one namespace.
- **Exported names:** a capital first letter exports an identifier (`cli.Run`); lowercase is package-private. No `public`/`private` keywords.
- **`internal/`:** packages under it can only be imported from inside the module — enforced by the compiler.
- **Errors are values:** functions return `error` instead of throwing; the caller checks `if err != nil`. Panics are for bugs (bad index, nil dereference), not for expected failures.
- **Implicit interfaces:** a type satisfies an interface by having its methods; no `implements`. `*cli.Error` is an `error` because it has `Error() string`.
- **Wrapping:** `fmt.Errorf("…: %w", err)` wraps; `Unwrap()` exposes the cause, like `getCause()`. `errors.Is` matches a value in the chain, `errors.As` finds a type (or interface) in it.
- **Pointer receivers:** `func (e *Error) Error()` — methods on the pointer, since errors are created as `&Error{…}`.
- **Zero values:** every type has one (`""`, `0`, `false`, `nil`); an unset struct field is its zero value, so there is nothing to initialise.
- **Value semantics:** structs are copied on assignment, pass and return, so building one in a local variable is invisible to everyone else. Slices, maps and pointers share their data — that is where mutation needs care.
- **No ternary, few expressions:** `if` is a statement; small pure helper functions replace conditional expressions.
- **Slices:** `append` returns the new slice (assign it back); `s[1:]` is a view without copying; `s...` spreads a slice into variadic arguments. A nil slice is a fine empty list.
- **No negative indexing:** an out-of-range index panics at run time.
- **Maps:** a nil map can be read but panics on write — create with `map[K]V{}`. `v, ok := m[k]` (comma-ok) tells a missing key from a zero value.
- **Closures:** function literals can assign to captured variables (no effectively-final rule). Function types are first class: a parameter of type `func(*flag.Flag)`.
- **`for` is the only loop:** `for { … break }` replaces `while`; `for _, x := range s` is for-each.
- **`if x := expr; cond {}`:** `x` is scoped to the `if`/`else`.
- **Struct tags:** `` `json:"ok,omitempty"` `` drives `encoding/json`, like Jackson annotations; `json:"-"` skips a field.
- **`any`:** alias for `interface{}`, holds any value.
- **`iota`:** auto-incrementing constant in a `const` block — Go's enum.
- **`init()`:** runs before `main`; used where a package variable's initializer would refer to itself (initialization cycle).
- **`//go:embed`:** compiles files into the binary as an `embed.FS`; only files under the package's own directory.
- **Build info:** `runtime/debug.ReadBuildInfo` gives the module version and VCS commit the toolchain stamped in — no ldflags needed since Go 1.24.
- **`flag` package:** a `FlagSet` per command; `ContinueOnError` returns errors instead of exiting; `Parse` stops at the first non-flag; `Visit` walks only flags that were set.
- **`io.Writer` injection:** pass streams as parameters instead of using `os.Stdout`, so tests run commands in-process with a `bytes.Buffer`.
- **Table-driven tests:** a slice of case structs plus `t.Run(name, …)` per case; a new case is one struct literal. Filesystem tests use `t.TempDir()`.
- **`gofmt`:** formatting is not a choice — unformatted code fails `make check`.
