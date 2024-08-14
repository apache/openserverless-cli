`executable` make a file executable: on Unix-like systems it will do a chmod u+x.
On Windows systems it will rename the file to .exe if needed.

```text
Usage:
    ops -executable <filename>
```
## Example

```bash
ops -executable kind
```