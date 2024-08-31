Show extension and MIME type of a file.
Supported types are documented [here](https://github.com/h2non/filetype?tab=readme-ov-file#supported-types)

```text
Usage:
    ops -filetype [-h] [-e] [-m] FILE
```

## Options

```
-h  shows this help
-e  show file standard extension
-m  show file mime type
```

## Examples

### File Mime type

```bash
ops -filetype -m `which ops`
```

This will output the ops executable type:
``application/x-mach-binary`` or `application/x-executable`

