# mia-help(1)

## NAME

mia-help - a manual page rendered as text

## DESCRIPTION

Prints a manual page to stdout as plain text. With no topic it prints mia(1).
A topic is a command (`mia help version`) or a page name
(`mia help config` for mia-config(5)). `mia <command> --help` prints the same
page as `mia help <command>`.

## OUTPUT

Text, not JSON. An error is printed to stderr.

## EXIT STATUS

`0` when the page was printed, `1` when there is no page for the topic.

## EXAMPLES

```
mia help
mia help man
mia commands --help
```

## SEE ALSO

mia(1), mia-man(1)
