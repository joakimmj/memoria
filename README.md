# Memoria

TUI (or CLI tool.. not decided yet) for notes and todos.

Programming language: Rust or Go

## Features

API
- add/edit/remove/complete todos
- add/edit/remove notes
- folder structure of notes as table of content
- view notes and todos (syntax highlighting? use som other tool for this?)
- edit notes with $EDITOR

UI
```
+------------------------------------------------------------------------------
|
|  <list of todos>                                                      |
|                                                                   scrollable
|  <list of completed todos> (default: hidden)                          |
|
+------------------------------------------------------------------------------
|  <keymap help>
+------------------------------------------------------------------------------
|  <keymap notes>
+------------------------------------------------------------------------------
|
|                                                                       |
|  <TOC notes>                                                      scrollable
|                                                                       |
|
+------------------------------------------------------------------------------
```

## Configuration
> `.config/memoria/mia.conf`

- directory for notes/todos (default: `~/.memoria/`)
- default edirot (default extracted from OS???)
- 

## Usage

```
$ mia -h

Memoria will default start TUI

```
