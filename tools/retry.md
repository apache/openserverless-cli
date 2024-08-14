```text
Usage:
    ops -retry [options] task [task options]
```

## Options

```text
-h, --help	Print help message
-t, --tries=#	Set max retries: Default 10
-m, --max=secs	Maximum time to run (set to 0 to disable): Default 60 seconds
-v, --verbose	Verbose output
```

## Example

Retry two times to get the ops action list

```bash
ops -retry -t 2 ops action list
```