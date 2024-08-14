`echoif` is a utility that echoes the value of `<a>` if the exit code of the previous command is 0,
echoes the value of `<b>` otherwise

```text
Usage:
    ops -echoif <a> <b>
```

## Example
```bash
  $( exit 1 ); ops -echoif "0" "1"
  1
```

or

```bash
  $( exit 0 ); ops -echoif "0" "1"
  0
```
