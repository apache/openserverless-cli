Check if a semver version A > semver version B.
Exits with 0 if greater, 1 otherwise.

```text
Usage:
    ops -needupdate <versionA> <versionB>
```

## Options

```
    -h, --help		 print this help info
```

## Examples

### Update is needed

```bash
ops -needupdate 1.0.1 1.0.0; echo $?
```

This will output:

```text
0
```

### Update is not needed

```bash
ops -needupdate 1.0.0 1.0.1; echo $?
```

This will output:

```text
1
```