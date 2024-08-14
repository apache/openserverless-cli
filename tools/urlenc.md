urlencode parameters using the default & separator (or a specific one using -s flag).
Optionally, encode the values retrieving them from environment variables.

```text
Usage:
    ops -urlenc [-e] [-s <string>] [parameters]
```

## Options
```
-e    Encode parameter values from environment variables
-h    Show help
-s string  Separator for concatenating the parameters (default "&")
```

## Examples

```bash
ops -urlenc a=1 b=2
```

This will output:

```text
a%3D1&b%3D2
```
