Join a relative path to the path from where `ops` was executed.
This command is useful when creating custom tasks ( e.g. an ops plugin).

```text
Usage:
    ops -opspath <path>
```

Options:

```
-h, --help  print this help info
```

## Examples

### You are executing in directory `/home/user/my/custom/dir`

```bash
ops -opspath my-file.txt
```

This will output:

```text
/home/user/my/custom/dir/my-file.txt
```
