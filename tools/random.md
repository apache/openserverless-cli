Generate random numbers, strings and uuids

```text
Usage:
    ops -random [options]
```

## Options

```
-h, --help  shows this help
-u, --uuid  generates a random uuid v4
--int  <max> [min] generates a random non-negative integer between min and max (default min=0)
--str  <len> [<characters>] generates an alphanumeric string of length <len> from the set of <characters> provided (default <characters>=a-zA-Z0-9)
```
 
## Examples

### Random uuid v4:

```bash
ops -random -u                 
```

This will output something like:

```text
5b2c45ef-7d15-4a15-84c6-29144393b621
```

### Random integer between max and min
```bash
ops -random --int 100 60
```

This will output something like:

```text
78
```