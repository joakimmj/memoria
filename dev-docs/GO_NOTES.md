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

## Phase 1.1a — slug

- **Bytes vs runes:** a `string` is UTF-8 bytes; `len("ø") == 2` and `s[0]` is a byte. A `rune` is one Unicode code point (an `int32`).
- **`for _, r := range s`:** decodes UTF-8 and yields runes — the Go equivalent of `codePoints()`, not `charAt`.
- **Rune vs string literals:** `'a'` is a rune, `"a"` a string; `string(r)` turns a rune into a string.
- **`strings.Builder`:** Go's `StringBuilder`. `s += x` in a loop copies the whole string each time (O(n²)), because strings are immutable. `Grow(n)` pre-sizes it.
- **Package-level `var` initializers:** run once, before `main`, in dependency order; every package is initialized eagerly at startup (Java initializes classes lazily on first use).
- **No `const` for composite values:** `const` is only for numbers, strings and booleans. A constant-like table is an unexported `var` that nothing writes; the package boundary stands in for `final`. Java's compile-time constants have the same limit — `static final Map` only blocks reassignment.
- **Doc comments:** a plain `//` comment directly above the declaration, starting with the identifier's name, in full sentences — no `/** */`, no `@param`. Expected on exported identifiers; optional on unexported ones.
- **`go doc`:** `go doc pkg` lists the package comment and exported signatures; `go doc pkg Name` shows one symbol's doc; `-all` shows everything, `-u` adds unexported names. In nvim, `K` (gopls hover) shows the same.

## Phase 1.1b — config

- **JSON into `any`:** objects become `map[string]any`, arrays `[]any`, numbers `float64`, `null` becomes `nil` — the `JsonNode` tree, versus decoding into a struct (the POJO).
- **Type assertion / type switch:** `m, ok := x.(map[string]any)` asks whether an `any` holds that type; `switch v := x.(type)` branches on it. Without `ok`, a wrong type panics.
- **Null vs absent in a map:** `v, ok := m[k]` — `ok && v == nil` is a JSON null, `!ok` is a missing key. A struct can't tell these apart: both leave the zero value.
- **Everything is passed by value, but some values are handles:** a map is a pointer to a hash table, a slice a (pointer, len, cap) header. Copying either shares the data — like a Java object reference. No `Map.copyOf`; deep-copy (`clone`) where a result must not share with its inputs.
- **Nil map reads are safe:** `m[k]` on a nil map returns the zero value and `false`; only writes panic. A failed `x.(map[string]any)` gives a nil map you can still index.
- **`append` can alias:** `append(s, x)` writes into `s`'s backing array when it has spare capacity, so two appends to the same `s` may overwrite each other.
- **`json.Unmarshal` is lenient by default:** unknown keys are ignored, missing keys keep the zero value, `null` into a non-pointer field is a no-op. `json.Decoder.DisallowUnknownFields` rejects unknown keys; nothing rejects missing ones — test for completeness instead.
- **Unmarshal into a filled struct merges:** present keys overwrite, absent keys stay — handy for layered config, unless `null` must mean something.
- **`reflect.TypeFor[T]()`:** a `reflect.Type` without a value; `typ.Fields()` iterates struct fields, `f.Tag.Get("json")` reads a tag.
- **`t.Setenv`:** sets an env var for one test and restores it after; the test can't run in parallel.
- **`go test -fullpath`:** prints full file paths in failures, so editors (`:make`, quickfix) can jump to them.

## Phase 1.1c — brains

- **Sentinel errors:** `var ErrX = errors.New("…")`, wrapped with `fmt.Errorf("(%s) %w", name, ErrX)`; callers test with `errors.Is`. Like catching a specific exception class, but the "type" is a value, and the message stays free-form.
- **Map iteration order is random:** deliberately, on every run. Stable output means sorting keys: `slices.Sorted(maps.Keys(m))`. `json.Marshal` of a map already sorts keys.
- **Struct embedding:** `struct{ brain.Brain; Exists bool }` promotes `Brain`'s fields and methods; `encoding/json` flattens them. Composition, not inheritance — no subtype relation.
- **Pointer receivers for types that change:** `func (r *Registry) …`. A value receiver gets a copy of the struct; a map field would still be shared, any other field would not.
- **Atomic file write:** `os.CreateTemp` in the target's directory, write, `Close`, `os.Rename` over the target. `defer os.Remove(tmp)` cleans up on every error path.
- **Existence check:** `os.Stat` plus `errors.Is(err, fs.ErrNotExist)` — there is no `exists()`. `os.MkdirAll` is a no-op on an existing folder.
- **`filepath.Abs`:** resolves against the working directory and cleans; `filepath.Base` is the last element.
- **`t.Chdir`:** changes the working directory for one test and restores it, like `t.Setenv`.
- **Mutation testing by hand:** break one rule, check a test fails, restore — proves a test pins the rule.

## Phase 1.1d — create-engram

- **`time` layouts:** Go formats with a reference date, `Mon Jan 2 15:04:05 MST 2006` — `t.Format("2006-01-02")`, not `yyyy-MM-dd`. Letters are literal: `t.Format("YYYY")` writes `YYYY`.
- **Inject `now time.Time`:** pass the time as a value; tests use a fixed `time.Date(...)`. No `Clock` interface needed, unlike Java's injected `Clock`.
- **`O_CREATE|O_EXCL`:** `os.OpenFile` fails with `fs.ErrExist` when the file exists — check and create in one step, no race (Java: `StandardOpenOption.CREATE_NEW`).
- **Labelled `continue`:** `next: for … { for … { continue next } }` restarts the outer loop from inside the inner one, as in Java.
- **Indexing a string gives bytes:** `format[i]` is a byte; copying byte by byte keeps UTF-8 intact when every token is ASCII. `len(s)` counts bytes too, so a byte column is just `len` arithmetic.
- **`strings.NewReplacer`:** replaces several patterns in one pass, so text it inserted is never matched again — no `%date%` expanding inside a title.
- **`strings.Cut`:** splits at the first separator into before, after and found — Go's `split(sep, 2)` without the slice.
- **`%%` in `fmt`:** a literal `%`; `"%body%"` in a format string would read `%b` as a verb.
- **Method values:** `expand := r.Replace` binds the receiver, giving a plain `func(string) string` (Java: `r::replace`).

## Phase 1.2a — markdown sections

- **Sub-slices share the backing array:** `lines[a:b]` is a view, not a copy. `append(lines[:i], x)` overwrites `lines[i]` in the caller's slice when capacity allows — the classic aliasing bug. Java's `subList` is a view too, but `add` on it throws or shifts; Go just writes.
- **`slices.Concat(a, b, c)`:** allocates one new slice and copies the parts in — the safe way to splice without touching any input.
- **`slices.Clone`:** copies a slice; a clone of an empty, non-nil sub-slice is empty and non-nil, so "present but empty" stays distinct from `nil` ("absent").
- **`slices.IndexFunc` on a sub-slice:** the index is relative to the sub-slice; add the offset back (`open + 1 + i`).
- **`strings.Split` vs `SplitAfter`:** `Split` drops the separator and leaves a trailing `""` for a final newline, so `Join(Split(s))` is an exact inverse.
- **Closures returning closures:** `is := func(m string) func(string) bool { … }` builds predicates on the spot (Java: a method returning a `Predicate<String>`).

## Phase 1.2b — synapses config and the SYNAPSES block

- **`json.Unmarshaler`:** a type with `UnmarshalJSON([]byte) error` decodes itself; `encoding/json` finds the method by its signature, no `implements`. Like a Jackson custom deserializer.
- **`type plain T` inside `UnmarshalJSON`:** a new type with T's fields but none of its methods, so decoding into it does not call `UnmarshalJSON` again. Convert back with `T(p)`.
- **Pre-filled decode target:** `json.Unmarshal` leaves fields the JSON omits as they were — set defaults first, then decode over them.
- **Decoder options don't propagate:** `DisallowUnknownFields` on the outer decoder does not reach a custom `UnmarshalJSON`; the inner decode must be strict itself.
- **Maps are not comparable:** `m1 == m2` does not compile (only `== nil`); use `maps.Equal`, or `reflect.DeepEqual` for nested values.
- **`maps.Clone` is shallow:** copies the top level only; nested maps are shared. Enough when only the top level changes.
- **`maps.DeleteFunc`:** removes the entries a predicate picks — no delete-while-ranging worries.
- **Comma-ok on a map:** `v, in := m[k]` separates "present with zero/nil value" from "absent" — the null-vs-missing distinction of a JSON tree.
- **Set union of keys:** `slices.Concat` the key lists, `slices.Sort`, `slices.Compact` (drops adjacent duplicates). Go has no `Set` type; `map[string]bool` is the other way.

## Phase 1.2c1 — engram arguments

- **`filepath.Rel` + `filepath.IsLocal`:** "is this path inside that folder" without string prefixes — `/bb/x` is not inside `/b`, and `..x` is a valid name. Java: `dir.relativize(p)` then check for `..`.
- **`info.Mode().IsRegular()`:** a plain file, not a folder, symlink or device.
- **Shadowing with `:=`:** inside a block, `x, err := f()` declares a new `x` hiding the outer one; to assign the outer, declare `err` first and use `=`. Java refuses a redeclared local; Go accepts it silently.
- **`slices.DeleteFunc`:** filters a slice in place and returns it shortened — only safe on a slice you own.
- **Returning a value with an error:** by convention the value is meaningless when `err != nil`; callers check `err` first, so a zero or partial value is fine.

## Phase 1.2c2 — attach-synapse

- **Pointers in a map:** `map[string]*doc` lets `d.changed = true` change the one cached value. Map values are not addressable, so with `map[string]doc` the assignment `m[k].changed = true` does not compile. Java maps always hold references; Go maps hold values unless you store pointers.
- **`regexp` is RE2:** linear time, no backreferences or lookaround. `FindStringSubmatch` returns `[whole, group1, …]`, or `nil` on no match. Compile once at package level with `MustCompile`.
- **`maps.EqualFunc(a, b, slices.Equal)`:** compares maps whose values are not comparable with `==`, such as `map[string][]string`.
- **`os.CreateTemp` makes files 0600:** an atomic replace has to copy the original mode (`Chmod(info.Mode().Perm())`), or every write makes the file private.
- **Byte scanning with offsets:** `strings.IndexByte(s[i:], c)` returns an index relative to `s[i:]`; add `i` back, and check for `-1` before you do.
- **Named results as documentation:** `func f() (file string, ok bool)` names what is returned in the signature; still return explicitly.

## Phase 1.2d — the atlas

- **Arrays vs slices:** `sha256.Sum256` returns `[32]byte`, an array — a value whose length is part of its type. `sum[:]` slices it for functions taking `[]byte`.
- **`os.ReadDir` + `DirEntry.Type()`:** lists a folder without a `Stat` per entry and without following symlinks; `Type().IsRegular()` keeps plain files only.
- **`time.Time` in JSON:** encodes as RFC 3339. `.UTC().Truncate(time.Second)` drops the monotonic clock and sub-seconds, so `…T10:00:00Z` round-trips to an equal value.
- **nil vs empty in JSON:** a nil slice or map encodes as `null`, an empty one as `[]` / `{}`. Initialise what readers expect to be a list.
- **`for i := range n`:** ranging over an int (Go 1.22) counts 0…n-1 — `for i := range markdown.FrontmatterEnd(lines)`.
- **Import cycles are compile errors:** `atlas` imports `engram`, so `engram` cannot call `atlas`; the caller (the CLI) sequences the two. Java allows cyclic packages; Go forces the dependency to point one way.
- **Struct embedding in a private type:** `stateTask{Task; state string}` adds a field to a public type without changing it.

## Phase 1.2e — engrams, engram, tasks

- **Custom `flag.Value`:** any type with `String()` and `Set(string) error`; `fs.Var` registers it and every occurrence calls `Set` — how a repeatable flag collects values.
- **`cmp.Or` chains comparisons:** returns the first non-zero argument, so `cmp.Or(cmp.Compare(a.X, b.X), cmp.Compare(a.Y, b.Y))` sorts by X then Y — Java's `Comparator.comparing(...).thenComparing(...)`.
- **Reverse order by swapping arguments:** `b.T.Compare(a.T)` sorts newest first; no `reversed()`.
- **A private struct to sort by a hidden key:** embed the public type plus the key (`struct{ Task; modified time.Time }`), sort, then copy the embedded value out.
- **Ranging over a map literal:** `for k, v := range map[K]V{…}` loops over a few named groups in one body; order is random, which is fine when the result is sorted afterwards.
- **`new(T)`:** allocates a zero `T` and returns `*T` — `new([]string)` for a slice a callee appends to through the pointer.

## Phase 1.2f — check and rebuild

- **`errors.As` takes a pointer to the target:** `var e *Error; errors.As(err, &e)` walks the wrap chain and sets `e` on a match — an `instanceof` + cast over the whole chain, on a returned value rather than a thrown one.
- **An error carrying data:** a struct implementing `error` can hold anything (`Result any`); callers that care unwrap it with `errors.As`, the rest just see the message.
- **Pointer to a slice element:** `p := &s[i]` changes the element in place; `for _, p := range s` hands out copies, and changes to them are lost.
- **Unexported fields in an exported struct:** `Problem{…; field, target string}` keeps package-internal data next to the JSON fields; `encoding/json` ignores unexported fields.
