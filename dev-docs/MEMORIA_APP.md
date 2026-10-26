# memoria

A note-taking system for folders of plain markdown files, usable from a
terminal, any editor, or an agent. One Go binary, `mia`, owns the model —
brains, engrams, synapses, concepts, the atlas — and every read and write of
it. Everything else is a client of that binary.

## Contents

- [1. Overview](#1-overview)
- [2. Shape](#2-shape)
- [3. Brains](#3-brains)
- [4. Configuration](#4-configuration)
- [5. Filenames](#5-filenames)
- [6. Markdown the core reads and writes](#6-markdown-the-core-reads-and-writes)
- [7. The engram header](#7-the-engram-header)
- [8. Synapses](#8-synapses)
- [9. Value fields](#9-value-fields)
- [10. Concepts](#10-concepts)
- [11. The atlas](#11-the-atlas)
- [12. Creating engrams](#12-creating-engrams)
- [13. Finding engrams](#13-finding-engrams)
- [14. Tasks and the agenda](#14-tasks-and-the-agenda)
- [15. The generated index](#15-the-generated-index)
- [16. Renaming and deleting engrams](#16-renaming-and-deleting-engrams)
- [17. The CLI](#17-the-cli)
- [18. The viewer](#18-the-viewer)
- [19. The language server](#19-the-language-server)
- [20. Editor clients](#20-editor-clients)
- [21. Build order and installing](#21-build-order-and-installing)
- [22. Decisions](#22-decisions)
- [Later](#later)
- [Open questions](#open-questions)

---

## 1. Overview

### 1.1 Vocabulary

| Term | Meaning |
|---|---|
| **brain** | A folder of engrams — one collection of notes |
| **engram** | One note: a markdown file in a brain |
| **synapse** | A structural connection from one engram to another (`up`, `down`, …) |
| **concept** | Something engrams refer to that is not itself a file — a person, a tag, a project |
| **mia** | The short form of memoria: the binary's name and the prefix of its hidden files |

### 1.2 Principles

- **Plain markdown first.** A brain is readable without memoria — on GitHub, in
  any editor, with grep. Synapses are markdown links; tasks are GitHub
  checkboxes.
- **Readable over ID-based.** Filenames and concept keys are words, not IDs;
  a rename propagates, which is the cost of that.
- **Derived data is rebuildable.** The atlas can be deleted and regenerated
  from the files with no loss.
- **Config holds deltas.** An override file holds only what differs.
- **No hidden state.** Nothing lives outside the files except what is
  inherently per-machine: where brains are.
- **The core never asks.** Every operation answers a result, or an error with
  a code; whoever called it decides whether a person is asked. Prompts,
  pickers and confirmations belong to clients.

### 1.3 Verbs

Operations are named by what they do to their noun, one meaning per verb, each
with its inverse:

| Verb | Means | Applies to | Inverse |
|---|---|---|---|
| **Register** | record something that exists outside memoria | brain (a folder), concept (a mention the registry does not know) | Deregister |
| **Create** | make something new | engram (a file), concept (a registry entry) | Delete |
| **Attach** | put a relation on an engram | synapse, concept | Detach |
| **Edit** | change what an existing thing carries | a concept's `meta`, an engram's value field | — |
| **Alias** | give a concept another name it answers to | concept | — |
| **Rename** | give an engram a new filename, keeping what refers to it | engram | itself |
| **Toggle** | flip a task between its two states | task | itself |

Rebuild (regenerate a derived file: the atlas, the index) and Search keep one
meaning each. Reads are named for what they answer (`engrams`, `agenda`). A
new operation picks its verb from this table; one that fits none adds a row
here first.

---

## 2. Shape

```
memoria (Go)                     clients
└─ mia                           ├─ a shell, a script, a git hook, an agent → mia <command>
   ├─ mia <command>   CLI        ├─ a person in a terminal                → mia view
   ├─ mia view        viewer     ├─ any LSP editor                        → mia lsp
   └─ mia lsp         LSP        └─ an editor plugin (memoria.nvim)       → mia lsp + mia <command>
```

- **One core, three surfaces.** The CLI, the viewer and the language server call
  the same Go packages and none of them holds a rule of its own. A surface
  adds only what it is: JSON on stdout, screens, or LSP messages.
- **The files are the shared state.** The atlas refreshes on read (§11.2), so
  the CLI, the viewer, a running language server and a person in an editor can
  work in one brain at once; whichever reads next picks up what the others
  wrote.
- **Editor plugins are thin.** They run `mia lsp` and `mia <command>`, and keep
  only what a client owns: pickers, views, navigation history (§20).
- **The CLI is stateless; clients hold state.** Each `mia <command>` run
  starts from its arguments, the working directory and the files, and keeps
  nothing after it exits — no current brain, no history. What only makes
  sense while something is running — the brain being browsed, navigation
  history, which done tasks are shown — lives in the viewer or an editor
  client, for as long as it runs.

The core's packages, by responsibility, each created by the build step that
first needs it (§21):

| Package | Holds |
|---|---|
| `brain` | the registry, brain resolution (§3) |
| `config` | defaults, the tiers, the merge (§4) |
| `markdown` | every construct the core reads or writes (§6) |
| `slug` | the one `slugify` (§5) |
| `engram` | the header, synapses, value fields, creation, rename, delete (§7–9, §12, §16) |
| `concept` | the registry file and resolution (§10) |
| `atlas` | the index, refresh, check, repair (§11) |
| `query` | neighbours, last, search, agenda, the generated index (§13–15) |
| `cli`, `view`, `lsp` | the three surfaces (§17–19) |

---

## 3. Brains

A **brain** is a folder of engrams. A user may have several (`work`,
`personal`), wherever they like on disk.

### 3.1 Registry

Which brains exist and where is per machine, so it is not part of any config
that travels between machines:

- stored at `$XDG_DATA_HOME/memoria/brains.json` (§4.4)
- written only by `mia` (`register-brain`, `deregister-brain`), never by hand
- locations are absolute paths, so a brain resolves the same from any working
  directory

```json
{
  "personal": { "location": "/mnt/sync/notes" },
  "work": { "location": "/home/me/dev/notes/work" }
}
```

Registering a path that does not exist creates the folder. The name defaults
to the folder's basename; a name already registered is an error.
Deregistering removes the entry only and never touches the folder.

### 3.2 Resolving the brain

An operation that needs a brain and was not told which resolves it in this
order:

```
1. the brain named by the caller (--brain)
2. the registered brain whose folder holds the file the operation is about
3. the registered brain whose folder holds the working directory
4. otherwise an error, code "brain", listing the registered brains
```

Only what the caller can see decides: the arguments and where the command
runs, the way `git` finds its repository. There is no current brain in the
core, since the CLI keeps no state between runs (§2). A brain that is named but
not registered is an error, never a fallback. A client that holds a brain for
the length of a session (§18, §20) passes it as step 1.

### 3.3 Contents

Flat — engrams sit at the brain's root, no subfolders:

```
my-brain/
├── .mia_concepts.json   hidden, hand-editable concept registry (§10)
├── .mia_atlas.json      hidden, derived index (§11)
├── .mia_dna.json        hidden, optional brain config (§4)
├── index.md             optional, generated overview (§15)
├── 2026-08-01_project_x.md
└── java_optional_stream.md
```

A bare folder of `.md` files is a valid brain; every other file is optional
and made on demand. `index.md` is an ordinary engram once written. The
`.mia_*` files are hidden because they are memoria's, not notes. A brain's
`.gitignore` leaves out `.mia_atlas.json`, which is derived, and keeps the
other two.

---

## 4. Configuration

### 4.1 Tiers

Three tiers, merged in order, later over earlier:

```
1. built-in defaults
2. the user's config:   $XDG_CONFIG_HOME/memoria/config.json (§4.4)
3. the brain's config:  <brain>/.mia_dna.json — only what differs
```

Both files are JSON, in the same shape, so a key means the same wherever it
is set. JSON has no comments; what may go in each file is documented in
`mia-config(5)` (§17.4) rather than written into the files.

The tiers are `mia`'s alone. An editor client configures only what belongs to
the editor — its commands, its keys, its views — in the editor's own config,
and passes nothing into these tiers.

Every surface loads the same tiers, so an engram written from a shell, the
viewer or an editor has the same header and template.

### 4.2 Defaults

Grouped by what each part shapes:

```json
{
  "engrams": {
    "date_format": "YYYY-MM-DD",
    "concept_form": "slug",
    "filename": { "prefix": "date", "date_format": "YYYY-MM-DD" },
    "content_template": "# %title%\n\n%cursor%",
    "task_markers": { "not_done": ["[ ]"], "done": ["[x]", "[X]"] },
    "index": {
      "filename": "index.md",
      "sections": ["stats", "recent", "concepts", "roots", "not_done_tasks", "tasks"],
      "stats": ["total_engrams", "total_concepts", "not_done_tasks"]
    }
  },
  "synapses": {
    "down": { "inverse": "up", "list": true, "show_empty": true },
    "up": { "inverse": "down", "list": true, "show_empty": true }
  },
  "frontmatter": {
    "created": { "kind": "value", "list": false, "default": "%date%" },
    "tags": { "kind": "concept", "concept_type": "tag", "list": true }
  },
  "concepts": {
    "tag": { "fields": ["description"] }
  },
  "agenda": {
    "depth": null
  }
}
```

| Key | Meaning |
|---|---|
| `engrams.date_format` | the `%date%` placeholder, wherever it is rendered (§12) |
| `engrams.concept_form` | how a concept field writes a concept by default: `"slug"` or `"display_name"` (§7.2) |
| `engrams.filename.prefix` | `"date"`, `"concept"` or `"none"` (§5) |
| `engrams.filename.date_format` | the `date` prefix's format |
| `engrams.content_template` | the prose a new engram starts with (§12) |
| `engrams.task_markers` | which list markers are tasks, by state (§14.1) |
| `engrams.index` | the generated index (§15) |
| `synapses` | the synapse fields (§7.2) |
| `frontmatter` | the frontmatter fields (§7.2) |
| `concepts` | concept type schemas (§10.2) |
| `agenda.depth` | how far the engram agenda walks; `null` walks all of it (§14.3) |

Date format tokens: `YYYY`, `YY`, `MM`, `DD`, `HH`, `mm`, `ss`; anything else
is written literally. Keys are snake_case in every tier.

### 4.3 Merging

- **Maps merge by key; lists replace wholesale.** `task_markers.done = ["[x]"]`
  means exactly `[x]`.
- **`null` is not the same as absent.** A key left out follows the tier above.
  A key set to `null` is removed, and then:
  - an entry of `synapses`, `frontmatter` or `concepts` stays removed — the one
    way to drop a default field (`{ "synapses": { "up": null } }`);
  - any other setting falls back to the built-in default, skipping the user
    tier (`{ "engrams": { "date_format": null } }`). A setting the code reads
    is never missing.
- **A tier that cannot be read** — missing, unreadable, not valid JSON — is
  skipped; one that exists but is invalid is reported, naming the file.

`.mia_dna.json` holds only what differs, and nothing seeds it with the merged
config: a full copy would freeze the brain against later changes to the
defaults or the user's config, and hide what is actually special about it.

```json
{ "engrams": { "filename": { "prefix": "none" } } }
```

`mia config [--brain B]` answers the merged result, for clients and for a
person checking what applies; `mia config --defaults` answers the built-in
defaults alone, to copy the keys worth changing from. Nothing writes a config
file on install or first run: a missing `config.json` means the defaults.

### 4.4 Platforms and locations

Linux and macOS are supported. Both use the XDG locations, macOS included,
since that is where a terminal user's dotfiles live:

| File | Location | Default |
|---|---|---|
| User config | `$XDG_CONFIG_HOME/memoria/config.json` | `~/.config/memoria/config.json` |
| Brain registry | `$XDG_DATA_HOME/memoria/brains.json` | `~/.local/share/memoria/brains.json` |

The config is meant for dotfiles and travels between machines; the registry
holds absolute paths and stays on the machine it describes.

Windows comes last (§21). Supporting it would put the two files under
`%AppData%\memoria\` and compare brain locations case-insensitively, with
either separator. Three rules keep that cheap until then:

- paths are built and split with `filepath`, never with `/` in a string;
- the two locations come from one function, the only place that knows them;
- brain locations are compared through one helper, on cleaned absolute
  paths.

---

## 5. Filenames

### 5.1 Prefix modes

| `prefix` | Example | |
|---|---|---|
| `"date"` | `2026-08-01_project_x.md` | `filename.date_format` of the creation date |
| `"concept"` | `java_optional_stream.md` | a registered concept's key (§5.4) |
| `"none"` | `optional_stream.md` | no prefix |

Prefix and slug are joined with `_`. The slug comes from the title given at
creation; the title itself is kept as given for `%title%` (§12).

### 5.2 `slugify`

```
slugify(text):
  1. trim; transliterate through memoria's own Latin map
     (æ → ae, ø → o, å → a, ä → a, ö → o, ü → u, ß → ss, é → e, ñ → n, …)
  2. lowercase
  3. every run of anything outside [a-z0-9] → "_"
  4. strip "_" at either end
```

| Text | Slug |
|---|---|
| `Note about Java` | `note_about_java` |
| `  Alice   Smith  ` | `alice_smith` |
| `Mary-Jane Watson` | `mary_jane_watson` |
| `Ærlig Østers` | `aerlig_osters` |
| `proj/web` | `proj_web` |
| `Node.js`, `O'Brien` | `node_js`, `o_brien` |
| `C++`, `C#` | `c`, `c` |
| `東京`, `???` | empty |

The same function makes filenames and concept keys (§10.1), so a concept
prefix is its key verbatim and a `#slug` in prose is spelled the way the
filename is. The map is memoria's own, so the result is the same on every
system; a character it does not know is not transliterated, so it falls
outside `[a-z0-9]` and separates words like any other (`a東b` → `a_b`).
That is lossy on purpose:
`alice-smith` and `alice_smith` are one name. A title that slugifies to
nothing is the error `empty_slug`.

### 5.3 Identity

A filename is the engram's **identity** once created. Editing the title in
the document changes nothing on disk, so fixing a typo in a heading does not
rewrite every link to the engram. Renaming is a separate, explicit operation
(§16.1).

A filename already taken is the error `collision`; the core never invents a
suffix.

### 5.4 Concept prefixes

With `prefix = "concept"`, creating an engram requires a concept:

```
1. no concept given                     → error, code "concept"
2. resolve it (§10.3); unregistered     → error, code "concept"
3. its key is the prefix
4. it is also written into a concept field: every field accepting its type
   (one declaring that type, or one declaring none) is a candidate;
   the caller names one, or there is exactly one — otherwise code "concept_field"
```

The concept is always written into a field as well as the filename, so a
`java_` prefix is never a `java` the atlas cannot see. A concept is never
registered implicitly here: registering needs a type nobody gave.

---

## 6. Markdown the core reads and writes

The `markdown` package is the one place every construct is read and written,
so how a thing is written and how it is parsed cannot drift apart. It works
over lines and byte ranges; every write is a minimal edit to the lines it
concerns, and the rest of the file stays byte-for-byte as it was.

### 6.1 Frontmatter

A `---`-delimited block on line 1, closed by a later `---`. An opening
delimiter never closed is not frontmatter.

- Fields are scalars and single-level lists. A list is block style (`- item`
  lines under the key) or flow style on one line (`tags: [java, proj_web]`,
  `tags: []`); items may be double-quoted to hold a comma
  (`["Smith, Alice"]`).
- A value is a string or a list, decided by what follows the key: a flow list
  or `- ` lines is a list, `tags: java` a string, and a bare `tags:` the empty
  string. Readers that expect a list normalise: empty or missing → `[]`, a
  string → one item.
- Comments (`# …`) are skipped.
- Each field carries the rows it spans, which is what a write replaces.

**Writing one field** replaces that field's rows with one line, adds a missing
field before the closing delimiter, gives a document without frontmatter a
block, removes the field for no value, and writes the bare key for an empty
`list = false` value. Every other line is kept. A block list it rewrites
comes back flow-style. Writing refuses — with the parse error, the file
untouched — when the frontmatter cannot be read, or the field cannot be
rewritten whole (followed by indented lines the reader does not model, or
written twice).

**Formatting a value:** a flow list for a list; a string plain when it reads
back unambiguously, double-quoted otherwise — for `[`, `]`, `{`, `}`, `,`,
`:`, `#`, a quote, a line break, edge whitespace, a leading YAML indicator, or
the empty string.

### 6.2 List items and tasks

A list item is `prefix`, an optional `marker`, and `text`:
`- [x] Send follow-up` is `"- "`, `"[x]"`, `"Send follow-up"`. A bracketed
run is a marker only when whitespace or the end of the line follows it, so
`- [Note](note.md) matters` is an item holding a link. Writing is the exact
inverse of reading.

A list item is a **task** when its marker is one of `engrams.task_markers`;
its state is the key the marker is listed under, `not_done` or `done`.

### 6.3 Links and headings

- A link is `[text](path)`. Images (`![…](…)`) are not links. Reference-style
  links are not followed. Links inside code spans and fenced code are
  skipped.
- A heading is an ATX heading, `#` to `######` followed by a space; `#word`
  is a word, not a heading.

### 6.4 Fenced sections

The mechanism behind the `SYNAPSES` block (§7) and the generated index
(§15): content regenerated wholesale between fixed markers.

```
<!-- NAME -->
…body…
<!-- /NAME -->
```

- **get(lines, name)** answers the body; an empty section is `[]`, no section
  is none.
- **set(lines, name, body, at?)** replaces the body; with no section present
  it writes one, markers included, before row `at`, or at the end. The result
  depends only on `body`.
- Markers match trimmed lines; the section is the first opening marker and
  the first closing one after it; both must be present.

Every other edit of a section is get, change, set.

### 6.5 Concept mentions in prose

A **mention** is `#` followed by a run of `[a-z0-9_]`, starting after
whitespace or at the start of a line — `#alice_smith`. Code spans, fenced
code, link destinations, headings, frontmatter and the `SYNAPSES` block are
skipped.

---

## 7. The engram header

### 7.1 Layout

An engram's header has two halves: **frontmatter** for its concepts and
values, and a **`SYNAPSES` block** for its connections to other engrams. The
block is markdown, not YAML, so a synapse is a link that renders and works in
any viewer.

```markdown
---
created: 2026-08-01
participants: [Alice Smith, Pete Park]
tags: [java]
---
<!-- SYNAPSES -->
- **down:** [note1aa](note1aa.md), [note1ab](note1ab.md)
- **up:**
***
<!-- /SYNAPSES -->

# Some note heading
```

The header is laid out for vertical space:

- frontmatter lists are flow-style, one line per field however many values;
- synapse fields are a tight list, one item per field;
- the `***` separator is the block's last line, inside the markers, so it is
  regenerated with the block and never orphaned;
- fields are sorted by name, in frontmatter and in the block;
- one blank line separates the header from the prose, written only when there
  is a header.

The block sits directly under the frontmatter, or on row 1 without one.

### 7.2 Field config

**Synapse fields** (`synapses`) are engram-to-engram; a value is a link to
another engram.

| Property | Meaning |
|---|---|
| `list` | whether several values are allowed (default `true`) |
| `inverse` | the paired field kept in sync on the target, e.g. `up` ↔ `down` |
| `show_empty` | whether the field's line appears with no values (default `true`) |

**Frontmatter fields** (`frontmatter`) are one line each, and their `kind`
says what a value is:

| `kind` | A value is |
|---|---|
| `"concept"` | a concept, resolved against the registry (§10) |
| `"value"` | a plain string, resolved against nothing |

| Property | Meaning |
|---|---|
| `kind` | required |
| `list` | several values allowed (default `true`); `false` writes the value itself, `room: kitchen`, or a bare `room:` when empty |
| `concept_type` | concept fields: the one type the field takes; omitted, it takes every type |
| `concept_form` | concept fields: `"slug"` (the concept's key) or `"display_name"`; default `engrams.concept_form` |
| `default` | value fields: what creation writes when given nothing, with the template placeholders rendered (§12) |

- **`concept_type` is enforced when given.** Attaching a concept of another
  type is an error, and a client's picker for the field lists that type only.
  Without it the field is a catch-all.
- **`concept_form` is per field**, because the field is what one list on disk
  looks like; a list mixing forms is worse than either. `slug` writes
  `tags: [java, proj_web]`, spelled like a prose mention; `display_name` writes
  `participants: [Alice Smith]`. A value in the other form or an alias still
  resolves, and `check` reports it as `non_canonical_mention` (§11.4).
- **`created` is a value field with a default**, written once at creation and
  never again unless edited. There is no `updated`: the atlas's `modified`
  and the file's mtime answer that, and a second timestamp could disagree.
- `participants`, a `person`-typed concept field, is the obvious addition for
  a meeting brain; it is configured per brain rather than shipped.

The config says what memoria writes, not what is valid: fields added by hand
are kept and left alone.

### 7.3 `SYNAPSES` block rules

- One list item per field: `- **label:** item, item`, bullet `- `, no
  checkbox.
- An empty field is `- **label:**` with no trailing space.
- Every configured field is written, unless `show_empty = false` and empty.
- One line per field regardless of how many values.
- Values are comma-separated links. A link's text is the target's filename
  without `.md` — its identity, which never goes stale the way a copied title
  would: `[2026-08-01_project_x](2026-08-01_project_x.md)`.
- The last line is `***`. A line that is not a field item is ignored when read
  and dropped on the next write; a marker on a field item is ignored and
  dropped.
- A field item the config does not name — added by hand, or removed from the
  config — is kept on every write, sorted in with the rest.
- Labels match `[A-Za-z0-9_]+`.

**Writing** regenerates the block from the engram's synapses: configured
fields and kept hand-added ones, sorted. A block with no field line to show
is not written at all. The frontmatter is only read, never changed, by a
block write; frontmatter that cannot be read refuses the write, since a block
placed above an unreadable opening `---` would break it for every reader.

**Reading** answers each field's links in order.

---

## 8. Synapses

### 8.1 Attach

```
attach-synapse(source, field, target)   → { brain, source, field, target }
  1. the brain holds source; field is a configured synapse field;
     source ≠ target; both files exist
  2. add target to source's field, unless there
  3. field.inverse names a configured synapse field →
       add source to target's inverse field, unless there
  4. a list = false field gives up its old value, and that engram loses
     source from its inverse field
  5. read every touched file first; any unreadable frontmatter refuses before
     anything is written; then write each file that changed
  6. refresh the atlas
```

Idempotent: values already present are left alone, so attaching an existing
one-sided synapse writes only the missing inverse — which is also how
`rebuild --fix` repairs one (§11.5).

### 8.2 Detach

```
detach-synapse(source, field, target)   → { brain, source, field, target }
  1. field is configured, and holds target — otherwise an error
  2. remove target from source's field
  3. field.inverse is configured and the target file exists →
       remove source from target's inverse field
  4. as attach, steps 5–6
```

Detaching leaves no one-sided synapse. A target whose file is gone can still
be detached; only the source is written, which clears a `broken_synapse`
without a hand edit.

---

## 9. Value fields

```
edit-frontmatter(source, field, value)   → { brain, source, field, value }
  1. field is a configured frontmatter field of kind "value"
     (a concept field is attach-concept's, which resolves and type-checks)
  2. value ""  → cleared: [] for a list, a bare key otherwise
     list=false → replaced
     list=true  → appended, unless there
  3. write the field (§6.1); refresh the atlas
```

A field's `default` plays no part here: it is what creation writes, not what
clearing restores.

---

## 10. Concepts

Concepts are what engrams refer to without being files: people, tags,
projects.

### 10.1 The registry — `.mia_concepts.json`

```json
{
  "alice_smith": {
    "aliases": ["Alice", "AS"],
    "display_name": "Alice Smith",
    "meta": { "email": "alice@example.com", "role": "designer" },
    "note": "Met at conference 2024",
    "type": "person"
  },
  "java": {
    "display_name": "Java",
    "meta": { "description": "Java-related engrams" },
    "type": "tag"
  }
}
```

- **The key is a slug** (§5.2), derived from the display name at creation or
  given instead, for a name the slug loses (`C++` → `cpp`) or two people
  sharing one. It has no spaces, so prose names it bare: `#alice_smith`.
- `display_name` — required and unique: what clients show and what a
  `display_name` field writes. No display name is another concept's alias.
- `type` — required, and one the brain has: a type with a schema in
  `concepts` (§10.2), or one the registry already uses. A type cannot be
  coined by a typo.
- `aliases`, `note` — optional. `meta` — free-form, shaped by the type's
  schema.
- Written two-space indented with keys sorted, one per line: it is meant for
  hand-editing, and a problem pointing at an entry needs a row. Empty
  `aliases` and `meta` are left out.
- Near-duplicates collide on the slug (`Alice Smith`, `alice-smith`) and are
  refused rather than becoming two people.
- Changing a display name is cheap: the old name still resolves by slug or as
  an alias, and `rebuild --fix` rewrites it. Changing a key means rewriting
  every engram, mention and prefix that uses it.

### 10.2 Type schemas

```json
{
  "concepts": {
    "person": { "fields": ["email", "role", "org"] },
    "tag": { "fields": ["description", "color"] }
  }
}
```

A schema lists the `meta` keys a client asks for when a concept of that type is
created or edited, and is where the list of types comes from. It is advisory:
unknown `meta` keys are kept. Only `tag` ships.

### 10.3 Resolving a mention

```
resolve(mention):
  1. a key                    → that concept
  2. a display_name           → that concept
  3. an alias                 → that concept
  4. slugify(mention) is a key → that concept
  5. otherwise                → undeclared
```

Keys, display names and aliases are kept unique against each other, so steps
1–3 never answer two ways. Step 4 lets a hand-typed `alice smith` resolve with
no alias. It comes last because a slug is lossy: an exact display name `C++`
on `cpp` must win over a concept keyed `c`.

### 10.4 Operations

```
create-concept(display_name, type, meta?, slug?)   → concept
  key = slug or slugify(display_name); empty → error
  a key already registered, or a display name that already resolves → error
  type must be one the brain has

edit-concept(mention, type?, meta?)                → concept
  the mention must resolve to a registered concept
  meta "k=" removes k

alias-concept(mention, alias)                      → concept
  an alias that already resolves to another concept → error

attach-concept(source, field, concept)             → { brain, source, field, text }
  1. field is a configured concept field
  2. resolved → its type must be the field's concept_type, if it has one;
                text = its key or display_name, per the field's concept_form
     undeclared → text = the mention as given
  3. already in the field (anything resolving to the same concept) → nothing
     written; list = false → replaces
  4. write the field (§6.1); refresh the atlas

detach-concept(source, field, concept)             → { brain, source, field, removed }
  removes every value in the field that is, or resolves to, the concept;
  none there → error. Other fields and prose mentions are untouched.

register-concept(mention, type | alias_of)         → concept
  an undeclared mention becomes a concept: created with the mention as its
  display name, or made an alias of an existing one. Then every engram naming
  it has that field rewritten into its concept_form (the same repair
  rebuild --fix makes, limited to those engrams).

undeclared-concepts                                 → { name, count, engrams }[]
  frontmatter mentions that resolve to nothing, by how many engrams name
  them, then by name
```

`create-concept` refuses a duplicate and `edit-concept` refuses an unknown
concept, so a typo in either is an error rather than a silent merge or a new
person. `attach-concept` writes an undeclared name as given, the way a
hand-typed tag would be; registering is a separate, optional cleanup.

---

## 11. The atlas

Derived, disposable, rebuildable from the engrams and the registry at any
time.

### 11.1 `.mia_atlas.json`

```json
{
  "config": "<fingerprint>",
  "concept_registry": "<fingerprint>",
  "engrams": {
    "2026-08-01_project_x.md": {
      "title": "Project X",
      "tags": ["java"],
      "participants": ["Alice Smith"],
      "synapses": { "down": [], "up": ["2026-07-30_kickoff.md"] },
      "links": ["2026-07-31_meeting_notes.md"],
      "modified": "2026-08-01T10:00:00Z",
      "hash": "a1b2c3"
    }
  },
  "backlinks": { "2026-07-31_meeting_notes.md": ["2026-08-01_project_x.md"] },
  "concepts": { "java": ["2026-08-01_project_x.md", "java_optional_stream.md"] },
  "concepts_by_type": { "person": ["alice_smith"], "tag": ["java"] },
  "tasks": {
    "not_done": [{ "engram": "2026-08-01_project_x.md", "line": 12, "text": "Call Alice" }],
    "done": [{ "engram": "2026-07-31_meeting_notes.md", "line": 4, "text": "Send follow-up" }]
  }
}
```

- **Per engram:** `title` is the first heading, or the filename without `.md`;
  one list per concept field, as written; `synapses` as filenames; `links`,
  the engrams its body links to; `modified` and `hash`; and `error` when its
  frontmatter cannot be read. Value fields are not indexed.
- **`backlinks`** maps a filename to every engram linking it, by synapse or
  body link, whether the file exists or not.
- **`concepts`** maps a concept to the engrams naming it — in a concept field,
  or in prose (§6.5). A mention is keyed by the concept it resolves to, so an
  alias and a display name index under the key; an undeclared frontmatter
  mention is keyed as written. **A prose mention is indexed only when it
  resolves**: prose is full of `#` that names nothing (`#fff`, `#42`).
- **`concepts_by_type`** lists every registered concept by type, referenced or
  not.
- **`tasks`** lists every task line outside the `SYNAPSES` block, by state.
  Tasks carry no metadata — no dates, no priority — so they stay plain
  checkboxes readable anywhere.

### 11.2 Refresh on read

Every read of the atlas first brings it up to date: every top-level `.md` file
is hashed, new and changed ones parsed, vanished ones dropped, and the file
written only when something changed. No watcher is needed, so an engram
edited anywhere is picked up the same way.

Two fingerprints sit beside the per-file hashes:

- `config` — of `synapses`, `frontmatter` and `engrams.task_markers`. A change
  re-parses every engram.
- `concept_registry` — of `.mia_concepts.json`. A change re-derives what
  depends on resolution and re-parses nothing.

Two processes writing the atlas at once leave the later write, which is
harmless for a derived file.

### 11.3 Rebuild

`rebuild` ignores every stored hash — the same result as deleting the atlas
first — and answers the atlas's problems. `check` is the same without writing
any engram.

### 11.4 Problems

Each problem has a `kind`, a `file` and a `line`:

| Kind | Meaning | Anchored on |
|---|---|---|
| `broken_synapse` | a synapse to a file that does not exist | the field's line |
| `missing_inverse` | a synapse whose inverse is not on the target | the field's line |
| `broken_link` | a body link to a file that does not exist | the link's line |
| `unreadable_frontmatter` | frontmatter that cannot be read | line 1 |
| `undeclared_concept` | a concept field value that resolves to nothing | the field's line |
| `wrong_concept_type` | a concept of another type than its field takes | the field's line |
| `non_canonical_mention` | a value that resolves but is not in its field's `concept_form` | the field's line |
| `orphaned_concept` | a registered concept no engram names | its key in `.mia_concepts.json` |

The four concept kinds are reported **only once the registry holds a
concept**: a brain that has declared nothing is not told everything it writes
is undeclared. `undeclared_concept` is a frontmatter question only. An alias
in use counts as a reference for `orphaned_concept`.

### 11.5 Repair

`rebuild --fix` first:

1. writes every missing inverse (attach, §8.1);
2. rewrites every `non_canonical_mention` into its field's form, merging two
   values for one concept;
3. regenerates every existing `SYNAPSES` block from config, which adds a line
   for a synapse field added since the engram was written.

Engrams without a block are left alone. The other problems have no repair:
registering needs a type nobody can invent, moving a concept between fields is
the writer's choice, deleting an entry or a link is destructive. They stay in
the report.

---

## 12. Creating engrams

```
create-engram(brain, title?, fields?, body?, concept?, concept_field?, reuse_if_exists?)
                                    → { path, cursor?, created }
  1. the filename (§5): prefix + slugify(title)
  2. header:
       frontmatter: every frontmatter field, sorted — given values from
       fields, else a value field's rendered default, else empty;
       no frontmatter fields → no frontmatter
       then the SYNAPSES block from config (§7.3); none to show → no block
  3. prose = content_template rendered
  4. content = header + (a blank line, when there is a header) + prose
  5. cursor = where %cursor% was; body, if given, goes there
  6. write the file; refresh the atlas
```

- `fields` maps a configured frontmatter field to its values; an unconfigured
  one is an error. A concept value is resolved, type-checked and written in
  the field's form, as `attach-concept` would.
- **The header is always generated from config**, never from a template, so a
  new engram is exactly what a rebuild would produce. `content_template`
  covers the prose only.
- **Errors with codes:** `empty_slug`, `collision`, `prefix` (an unknown
  `filename.prefix`), `concept`, `concept_field` (§5.4).
- **`reuse_if_exists`** changes only `collision`: the existing file is the
  answer, `created = false`, and nothing is written. The filename follows
  from the title, so this is one check, not a search. In a `date` brain a
  fixed title such as `today` makes this a page per day.
- **A missing title** is an error, except in a `date` brain, where the
  filename is the date alone.

| Placeholder | Expands to |
|---|---|
| `%title%` | the title as given |
| `%date%` | today, in `engrams.date_format` |
| `%cursor%` | where writing starts; removed from the file |

The resulting file with the defaults:

```markdown
---
created: 2026-08-01
tags: []
---
<!-- SYNAPSES -->
- **down:**
- **up:**
***
<!-- /SYNAPSES -->

# Project X

```

---

## 13. Finding engrams

### 13.1 Engrams and one engram

- `engrams [--concept C]` — every engram: filename, title, concept fields,
  `modified`; with `--concept`, those naming `C`, resolved first, so any
  spelling finds it.
- `engram <file>` — its atlas entry, its backlinks, its content, and the
  brain's two date formats, for reading the engram's date: the filename
  prefix in `filename.date_format`, else `created` in `date_format`.
- `frontmatter <file>` — every configured frontmatter field, sorted by name:
  `name`, `kind`, `list`, and for concept fields `concept_type` and
  `concept_form`, with its values as written (`[]` or `""` when empty); a
  concept field adds each value's resolution, `{ text, key, display_name }`.
  Read from the file, not the atlas, since value fields are not indexed.
  Fields in the file the config does not name are left out.

### 13.2 Neighbours

Everything one engram connects to:

```
neighbours(file)   → { synapses, backlinks, concepts }
  synapses  = per synapse field, in name order: { field, path, title }
  backlinks = engrams linking here that are not among its synapses: { path, title }
  concepts  = per concept field, in name order: { field, key, engrams },
              key as resolved (the text as written when undeclared),
              engrams = the other engrams naming it
```

This is what a navigation view shows: synapses by field, backlinks, and each
concept as `#key` with the engrams that share it.

### 13.3 Last engram

`last` answers the engram with the newest `modified`.

### 13.4 Search

```
search(query, fields?)   → { path, title, score }[]
```

Fuzzy matching over each engram's title, filename and concept fields, from
the atlas; a concept value matches by what it is written as and by its
concept's display name, so `alice` finds `participants: [Alice Smith]` and
`tags: [alice_smith]` alike. `fields` narrows which are matched. Best first.
No body search: the atlas holds no bodies.

### 13.5 Resolving a word

```
resolve-word(text)   → { concept } | { engram } | {}
```

What a word should become when a person turns it into a reference: a
registered concept (resolved, §10.3) or an engram whose title or slug
matches. A client then writes `#key` for a concept, `[text](path)` for an
engram, or offers to create an engram with the word as its title.

---

## 14. Tasks and the agenda

### 14.1 Tasks

`tasks [--state not_done|done] [--engram <file>]` — every task, by state; with
`--engram`, one engram's (an unknown file is an error).

### 14.2 Toggling a task

```
toggle-task(file, line)   → { file, line, state }
  1. the line must still be the task the atlas indexed — otherwise an error,
     code "stale", and that engram is re-indexed
  2. not_done → the first done marker; done → the first not_done marker
  3. rewrite that line only; refresh the atlas
```

### 14.3 The engram agenda

An agenda for one engram and everything connected to it:

```
agenda(file, depth?)   → { chain, tasks, done, total }
  1. depth = depth or agenda.depth; none walks the whole chain
  2. breadth-first from file over every synapse field, both directions,
     with a visited set — each engram once, at its shortest distance, so a
     cycle ends the walk
  3. chain = [{ distance, engrams: [{ path, title }] }], distance 0 the file;
     within a distance by date where the engram has one, undated last
  4. tasks = every task of every engram in the chain, in chain order, each
     with its engram; done and total count all of them
```

A client renders it as a chain section and a task list:

```markdown
# Agenda — Weekly sync

## Chain

- Weekly sync
- 1 step
  - Project kickoff
  - Follow-up
- 2 steps
  - Q4 planning

## Tasks

Done: 3 / 7

- [ ] Agree on scope — Weekly sync
- [ ] Send budget draft — Project kickoff
```

The brain agenda is `tasks`, grouped by engram, most recently modified first.

---

## 15. The generated index

`rebuild-index` writes `index.md` at the brain's root: a plain-markdown
overview, exportable with any markdown tool.

Each entry of `engrams.index.sections` is a fenced section (§6.4) named
`INDEX:<NAME>`, written in config order on first generation; text outside the
sections, such as a hand-written introduction, is kept. A section not listed
is left alone. `engrams.index.stats` lists the bullets inside `INDEX:STATS`.

| Section | Holds |
|---|---|
| `stats` | one bullet per `stats` entry: engrams, concepts, open tasks |
| `recent` | engrams by `modified`, newest first, linked |
| `concepts` | display names by type |
| `roots` | engrams with an empty `up` field — the closest a flat brain has to a table of contents |
| `not_done_tasks` | open tasks, each linking its engram |
| `tasks` | every task, each linking its engram |

```markdown
# Work

<!-- INDEX:STATS -->
- Engrams: 42
- Concepts: 18
- Open tasks: 5
<!-- /INDEX:STATS -->

<!-- INDEX:NOT_DONE_TASKS -->
- [ ] Call Alice about budget ([Project X](2026-08-01_project_x.md))
<!-- /INDEX:NOT_DONE_TASKS -->
```

`index.md` is otherwise an ordinary engram.

---

## 16. Renaming and deleting engrams

Both change files beyond the one named. Rename repairs every reference it
makes stale, so it runs as asked; delete breaks them, so it refuses until the
caller confirms with `force`.

### 16.1 Rename

```
rename-engram(file, title)   → { path, rewritten }
  1. new filename: the original's prefix kept (a leading date in
     filename.date_format; the longest registered concept key followed by _),
     new slug from title
  2. empty_slug / collision as for create
  3. rename the file; rewrite every synapse to it; rewrite body links to
     it, brain-wide, best effort
  4. refresh the atlas
```

Body links are rewritten best effort: links written the normal way are
caught, unusual hand-written ones may not be, and `check` reports what was
missed as `broken_link`.

### 16.2 Delete

```
delete-engram(file, force?)   → { referencing }
  1. referencing = engrams with a synapse to it, or a body link
  2. referencing non-empty and not force → error, code "referenced", listing them
  3. remove it from every synapse field naming it
  4. delete the file; refresh the atlas
```

Synapses are cleaned up, undoing what attach kept in sync. Body links are left
as they are and reported as `broken_link`.

---

## 17. The CLI

`mia <command>` is the core for anything that can run a command and read
JSON: a script, a git hook, an agent, an editor plugin. A brain is readable
without memoria; what needs memoria is writing a correct one — a header that
matches the config, a synapse with its inverse, an atlas that knows.

### 17.1 Invocation and output

```
mia <command> [--brain <name>] [args]
mia [command] --help
```

One JSON object on stdout, exit status to match:

```json
{ "ok": true, "result": { "path": "/home/me/notes/work/2026-08-01_project_x.md", "created": true } }
{ "ok": false, "error": "(work) 2026-08-01_project_x.md already exists", "code": "collision" }
```

`0` for `ok`, `1` otherwise. Errors name their brain, `(work) …`. `code` is
present wherever a caller can act on the failure. Diagnostics go to stderr,
so stdout is always one object. `--help` prints usage as text and loads no
config, so it works when nothing else does. `<file>` arguments are filenames
in the brain, or paths, which also decide the brain (§3.2).

Two codes belong to the CLI itself, not to any operation:

| Code | Meaning |
|---|---|
| `usage` | the command line is wrong: no command, an unknown command, a missing or extra argument, an unknown flag, a missing required flag |
| `not_found` | a named thing does not exist, such as a help topic |

`help` is the one command that does not answer JSON; when it fails, the error
goes to stderr, stdout stays empty, and the exit status is `1`.

### 17.2 Commands

An ordered list: `commands` prints it as a schema, `--help` reads the same
table, and their order is stable.

| Command | Answers |
|---|---|
| `brains` | name, location, whether it exists, whether it has `.mia_dna.json` |
| `register-brain <path> [--name N]` | the brain registered |
| `deregister-brain <name>` | the brain removed |
| `config [--defaults]` | the merged config, or with `--defaults` the built-in defaults alone (§4) |
| `engrams [--concept C]` | §13.1 |
| `engram <file>` | §13.1 |
| `frontmatter <file>` | §13.1 |
| `neighbours <file>` | §13.2 |
| `last` | §13.3 |
| `search --query Q [--fields F]` | §13.4 |
| `resolve-word <text>` | §13.5 |
| `tasks [--state S] [--engram <file>]` | §14.1 |
| `agenda <file> [--depth N]` | §14.3 |
| `concepts [--type T]` | per concept: key, display name, type, aliases, note, meta, engrams naming it |
| `undeclared-concepts` | §10.4 |
| `check` | the problems (§11.4) |
| `rebuild [--fix]` | the problems left after it |
| `rebuild-index` | the path written |
| `create-engram [--title T] [--field k=v,…] [--body -] [--concept C] [--concept-field F] [--reuse-if-exists]` | `{ path, cursor?, created }` |
| `attach-synapse <source> <field> <target>` | §8.1 |
| `detach-synapse <source> <field> <target>` | §8.2 |
| `edit-frontmatter <source> <field> <value>` | §9 |
| `create-concept --name N --type T [--slug S] [--meta k=v]` | §10.4 |
| `edit-concept --name N [--type T] [--meta k=v]` | §10.4 |
| `alias-concept --name N --alias A` | §10.4 |
| `register-concept --mention M (--type T \| --alias-of K)` | §10.4 |
| `attach-concept <source> <field> <concept>` | §10.4 |
| `detach-concept <source> <field> <concept>` | §10.4 |
| `toggle-task <file> <line>` | §14.2 |
| `rename-engram <file> --title T` | §16.1 |
| `delete-engram <file> [--force]` | §16.2 |
| `version` | `{ version, commit }`, set at build time from the git tag |
| `help [topic]` | a manual page rendered as text (§17.4); not JSON |
| `man --dir <dir>` | the manual pages written as roff into `<dir>` (§17.4) |
| `commands` | every command with its arguments, as a schema |

- `--field tags=java,streams` fills one field, once per field; values are
  comma-split. `--meta k=v` splits at the first `=` and is not comma-split.
  `--meta k=` removes a key.
- `--body -` reads the body from stdin; `--body <text>` takes it as given.

### 17.3 Running beside others

`mia` writes files, and holds nothing between runs. An editor with the file
open reloads it, or warns when it has unsaved changes; the language server
(§19) sends edits to open documents through the editor instead of writing
behind it. The atlas refreshes on every read, so every process sees every
other's writes.

### 17.4 Help and manual pages

The documentation is written once, as markdown, and compiled into the binary,
so it is there after a bare `go install`:

```
docs/
├── mia.1.md                  overview, global flags, output, exit status, the command list
├── mia-<command>.1.md        one page per command: mia-create-engram.1.md, …
├── mia-view.1.md             the viewer (§18)
├── mia-lsp.1.md              the language server (§19)
├── mia-config.5.md           config.json and .mia_dna.json, every key
├── mia-engram.5.md           the engram file: frontmatter, the SYNAPSES block, tasks, mentions
├── mia-concepts.5.md         .mia_concepts.json
└── memoria.7.md              brains, engrams, synapses, concepts, the atlas
```

- **Pages follow the man-page sections**, in this order, leaving out those
  that do not apply: NAME, SYNOPSIS, DESCRIPTION, COMMANDS, OPTIONS, OUTPUT,
  ERRORS, EXIT STATUS, EXAMPLES, SEE ALSO. COMMANDS is `mia(1)`'s alone:
  the command list.
- **`mia help [topic]`** renders a page in the terminal: `mia help` is
  `mia(1)`, `mia help create-engram` is `mia-create-engram(1)`,
  `mia help config` is `mia-config(5)`. `mia <command> --help` is the same
  page.
- **`mia man --dir <dir>`** writes every page as roff
  (`<dir>/man1/mia-create-engram.1`, …), converted inside the binary, so
  `man mia-create-engram` works once `<dir>` is on `MANPATH`. Release
  archives and packages ship the same pages.
- **SYNOPSIS, OPTIONS and COMMANDS come from the command table** (§17.2),
  the same one `commands` prints; only the prose is written by hand, so a
  page cannot disagree with the arguments a command takes.
- **A test holds the docs to the code:** every command in the table has a
  page, and every config key a line in `mia-config(5)`. A feature without its
  documentation fails, so the two land in the same commit.

---

## 18. The viewer

`mia view [file]` is for reading and moving around a brain in a terminal. It
does not write prose: it hands the terminal to `$VISUAL` or `$EDITOR` on the
engram shown and takes it back when the editor exits, then shows the engram
as it now is.

### 18.1 Opening

| `[file]` | Opens |
|---|---|
| nothing | the engram picker of the brain resolved (§3.2); with none resolved, the brain picker first |
| a path to an engram | that engram; its folder decides the brain |
| a filename, with or without `.md` | that engram in the brain resolved (§3.2) |
| a registered brain's folder | that brain's engram picker |

An engram is a top-level `.md` file in a registered brain's folder; anything
else is an error before the screen opens, on stderr with a non-zero exit like
any other command:

| Case | Code |
|---|---|
| a `.md` file outside every registered brain | `not_in_brain`, with the hint to `register-brain` its folder |
| a `.md` file in a subfolder of a brain | `not_an_engram` |
| a filename with no brain resolved | `brain` |
| a file that does not exist, or is not `.md` | `not_found` |

The viewer creates nothing: a missing engram is `not_found`, and creating is
`create-engram`'s or the editor's.

### 18.2 Screens

- **Brains and engrams** — pickers, with search (§13.4).
- **The engram view** — frontmatter, synapses, tasks and body each in its own
  part; the body rendered as markdown.
- **Navigate** — the current engram's neighbours (§13.2); a concept opens the
  engrams sharing it.
- **The engram agenda** (§14.3) and **the brain agenda**, jumping to a task's
  engram.

Each screen lists its own keys.

### 18.3 What it changes

One thing: **toggling a task** (§14.2), in the engram view and both agendas.
Everything else — prose, synapses, concepts, fields — is the editor's, one
key away.

### 18.4 Session state

While it runs, the viewer holds what the CLI does not (§2): the brain being
browsed, the navigation history that back returns along, and whether done
tasks are shown. Nothing of it is written anywhere; the next `mia view` starts
from its arguments again. It calls the core packages directly rather than
running `mia` per read, and refreshes the atlas before each screen, so changes
made in the editor or by another process show up.

---

## 19. The language server

`mia lsp` is a long-running server over stdio for any editor with LSP
support. It serves markdown files inside a registered brain, keeps each
brain's atlas in memory, and refreshes it as files change.

| LSP | memoria |
|---|---|
| Completion | engram targets in the `SYNAPSES` block; concepts of the field's type in a concept field; concept keys after `#` in prose; engram paths in a link |
| Go to definition | a synapse or link → its engram; a `#key` → its entry in `.mia_concepts.json` |
| References | on an engram's title → its backlinks; on a concept → every engram naming it |
| Hover | an engram's title and fields; a concept's display name, type and meta |
| Rename | on an engram → `rename-engram`, as one workspace edit |
| Diagnostics | the problems (§11.4), on their lines |
| Code actions | the repairs of `rebuild --fix`; register an undeclared concept; turn a word into a reference (§13.5); toggle a task |
| Workspace symbols | engrams by title, through search |

Edits to open documents go out as workspace edits, so the editor applies them
to what it holds; files that are not open are written to disk. Every CLI read
is also available as `workspace/executeCommand`, answering the same JSON, so
an editor plugin needs no second process for them. The server reports its
version in `serverInfo`, the same as `mia version`.

---

## 20. Editor clients

An editor integration is LSP plus whatever LSP cannot express, which a thin
plugin adds by calling `mia`. What a client owns, since none of it is a rule
about the brain:

- **Navigation history** — back across engram jumps, per session.
- **The active brain** — a session's default, passed as `--brain`.
- **Pickers and views** — choosing a brain, a field, a concept; the navigate
  picker over `neighbours`; agenda views over `tasks` and `agenda`.
- **Writing at a cursor** — inserting `#key` or a link from `resolve-word`,
  over LSP completion and code actions where the editor has them.

A client that needs a minimum `mia` checks `mia version` (or the server's
`serverInfo`) when it starts, and says which version it needs rather than
failing on the first missing command.

**memoria.nvim** is the Neovim client: `:Mia*` commands over `mia`, the
language server attached to markdown buffers in a brain, Neovim pickers and an
agenda buffer.

---

## 21. Build order and installing

The CLI is built first and completely; the viewer and the language server follow
on top of a finished core. Within the CLI, a feature comes once everything it
depends on exists, core before convenience:

**Phase 0 — the project**

The bare minimum of a working CLI, with nothing about brains yet: what every
later command plugs into.

- the Go module, `cmd/mia`, and the `cli` package (§2)
- the command table (§17.2) and argument parsing
- the output envelope and exit status (§17.1), stderr for diagnostics
- `version`, its value set at build time from the git tag
- `help`, `man --dir`, and `commands`
- `docs/` with `mia.1.md` and the pages for these commands (§17.4), and the
  test that every command has a page
- `go install …/cmd/mia@latest` working

**Phase 1 — the CLI**

1. **Brains and engrams** — the registry, config tiers, `slugify`, filenames
   (`date` and `none`), `create-engram`, the `markdown` package for
   frontmatter and fenced sections
2. **Synapses and the atlas** — the `SYNAPSES` block, attach, the atlas with
   refresh, `check`, `rebuild [--fix]`, `engrams`, `engram`, `tasks`
3. **Concepts** — the registry, resolution, concept and value fields,
   `concept_form`, concept prefixes, every concept command, `frontmatter`,
   `edit-frontmatter`, the concept problems and their repair
4. **Navigation** — `neighbours`, `last`, `create-engram --reuse-if-exists`
5. **Agenda** — `agenda`, `toggle-task`, `tasks --engram`
6. **Search and words** — `search`, `resolve-word`, prose mentions in the
   atlas
7. **Generated index** — `rebuild-index`
8. **Rename, delete and detach** — `rename-engram`, `delete-engram`,
   `detach-synapse`, `detach-concept`

**Phase 2 — the viewer**, over the finished core.

**Phase 3 — the language server**, over the finished core.

**Phase 4 — memoria.nvim** as a client of `mia`.

**Phase 5 — Windows** (§4.4).

Every step ships with its tests and its manual pages (§17.4), creates the
packages of §2 it first needs and no others, and is `go vet` and `gofmt`
clean.

**Installing.** `go install github.com/joakimmj/memoria/cmd/mia@latest` from
the first release; prebuilt binaries for Linux and macOS attached to tagged
releases to follow.

---

## 22. Decisions

- **One binary owns the model; clients own interaction.** A rule lives in one
  place, and a new client — another editor, a web view — adds screens, not
  logic.
- **The core never prompts.** Errors carry codes, so a client can ask and a
  script can decide.
- **The CLI before the viewer and LSP.** Both are built on the same packages, and
  a complete, tested CLI is what they are built on.
- **No implicit suffix on a filename collision.** Naming stays the user's: the
  slug comes from the title, never invented.
- **Synapses are a markdown block, not YAML** — links render and work in any
  viewer, and the block is regenerated whole between fixed markers. The
  markers are constants, not settings.
- **The header is laid out for vertical space** — flow lists, one item per
  synapse field, the separator inside the block, fields sorted by name.
- **A concept prefix is also a concept field value** — never cosmetic, so
  nothing visible in a filename is invisible to the atlas.
- **Tasks carry no metadata** — plain GitHub checkboxes. Which markers count
  is configurable, so an existing brain's convention is not ignored.
- **One strict `slugify` for filenames and concept keys** — the same on every
  system, safe everywhere a filename or a `#word` goes.
- **Exact mentions resolve before slugified ones** — a slug is lossy.
- **Concept keys are slugs; `display_name` is the natural name** — a bare
  `#alice_smith` needs no brackets, and near-duplicates collide.
- **`concept_form` per field, `slug` by default; `concept_type` optional and
  enforced** — the field knows what its list on disk looks like, and a typed
  field keeps people out of `tags`.
- **`created` is a value field with a default; no `updated`.**
- **A concept in prose is `#key`, not a link** — a concept is not a file.
  Prose mentions are indexed only when they resolve.
- **Concepts are reported only once one is declared.**
- **Opening by title is `create-engram --reuse-if-exists`**, not an operation
  of its own; a page per day is a fixed title in a `date` brain.
- **Rename rewrites body links best effort; delete does not touch them** — a
  rename changes identity, a delete's dangling links are reported.
- **Delete waits for `force`; rename does not** — a delete breaks what refers
  to the engram, so a caller confirms, and a script that passes `force` has
  decided; a rename repairs every reference it makes stale.
- **One config format** — the user's config and a brain's are both JSON, the
  same shape; documentation lives in the manual pages, not in comments.
- **Documentation is markdown compiled into the binary** (§17.4) — one page
  per command, readable as `mia help` with nothing installed and as `man`
  pages once written out; a test fails a feature that lands without its page.
- **Editors configure the editor** — a client's settings are its own; the
  model's config is `mia`'s tiers only.
- **The engram agenda walks breadth-first, grouped by distance** — each engram
  once, cycles end the walk, and `depth` caps a densely linked brain.
- **No current brain in the core** (§3.2) — the brain comes from the
  arguments or the working directory, never from an earlier run; a brain held
  for a session is a client's.
- **The CLI is stateless; clients hold session state** (§2) — every run can be
  understood from its command line alone, and what only matters while
  something runs — history, the brain being browsed — dies with it.
- **The viewer views** (§18) — it reads, navigates and toggles tasks; writing
  prose, creating engrams and editing structure are the editor's and the
  CLI's.
- **The registry is per machine and written only by `mia`** — locations do not
  travel between machines; the config does.
- **The command list is ordered** — `commands` answers the same twice, and
  anything generated from it inherits the order.

---

## Later

Wanted, not scheduled:

- **Structural edits in the viewer** — attaching and detaching synapses and
  concepts from the engram view, with pickers for the target, the field and
  the concept.
- **Quick capture** — `append`, a line added to a named section of an engram
  from a template (`- [ ] %input%`) without opening it.
- **MCP server** — `mia mcp`, a stdio server whose tools are generated from
  the command table (§17.2), so the CLI and MCP cannot drift. Long-running, so
  it holds the atlas in memory between calls. The CLI already serves any agent
  with a shell; MCP adds discovery for clients without one.

## Open questions

1. **Packaging beyond `go install` and release binaries** — Homebrew, distro
   packages, and whether editor clients download a matching binary.
