Check if a value is valid according to the given constraints.
If -e is specified, the value is retrieved from the environment variable with the given name.

```text
Usage:
ops -validate [-e] [-m | -n | -r <regex>] <value> [<message>]
```

## Options:

```
-e    Retrieve value from the environment variable with the given name.
-h    Print this help message.
-m    Check if the value is a valid email address.
-n    Check if the value is a number.
-r string Check if the value matches the given regular expression.
```
