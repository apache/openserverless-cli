# Tool: datefmt

Print date with different formats. If no time stamp or date strings are given, uses current time

```text
Usage:
  ops -datefmt [options] [arguments]
```

## Options

```
    -h, --help		 print this help info
    -t, --timestamp	 unix timestamp to format (default: current time)
    -s, --str 	  	 date string to format
    --if			 input format to use with input date string (via --str)
    -f, --of		 output format to use (default: UnixDate)
```

Possible formats (they follows the standard naming of go time formats, with the addition of 'Millisecond' and 'ms'):

- Layout
- ANSIC
- UnixDate
- RubyDate
- RFC822
- RFC822Z
- RFC850
- RFC1123
- RFC1123Z
- RFC3339
- RFC3339Nano
- Kitchen
- Stamp
- StampMilli
- StampMicro
- StampNano
- DateTime
- DateOnly
- TimeOnly
- Milliseconds
- ms

## Example

```bash
$ ops -datefmt -f DateTime
2024-08-11 03:00:34
```
