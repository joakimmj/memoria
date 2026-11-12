# mia-deregister-brain(1)

## NAME

mia-deregister-brain - remove a brain from the registry, leaving its folder alone

## DESCRIPTION

Removes the brain named `<name>` from the registry. Its folder and every file
in it stay where they are; registering the folder again brings the brain back.

## OUTPUT

```
{ "ok": true, "result": { "name": "work", "location": "/home/me/notes/work" } }
```

## ERRORS

- `not_found`: no brain has that name.

## EXIT STATUS

`0` when the brain was removed, `1` otherwise.

## SEE ALSO

mia-brains(1), mia-register-brain(1)
