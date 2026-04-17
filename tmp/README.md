# Memoria

A note taking tool/philosophy based on the Zettelkasten method, popularised by
sociologist Niklas Luhmann.

The idea is to create small, atomic notes — each with one idea written in your
own words and linking them together to build a network of connected knowledge.

## Implementation

My implementation of this follows these rules:
- flat structure for notes
- notes can be linked up and down
- notes can be tagged

Structure:
- you can have multiple archives (e.g. work, programming, etc.)
- each archive is a directory with flat structured notes
- all notes should have front matter with tags, up/down links

### Example

Structure:
```text
work/
   test-users.md
   roles.md
   services.md
java/
   optional.md
   optional-to-stream.md
```

`optional-to-stream.md`:
```md
---
tags:
  - java
  - null-safe
  - optional
up:
  - "[[optional.md]]"
down:
  - "[[haha]]"
  - "[[some_note.md]]"
---

# Optional to stream

When having optional of a List and you want to handle it, you can:

```java
public static List<Entry> getEntries(final SomeResponse response) {
    return Optional.ofNullable(response)
        .map(SomeResponse::entries)
        .stream()
        .flatMap(List::stream)
        .filter(Entry::valid)
        .toList();
}
```

[[some_folder/Hei.md]]


## Some quote

> quote

==Hhhh==

[[some_note]]

![[some_note]]
```

