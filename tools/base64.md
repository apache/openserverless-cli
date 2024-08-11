# Tool base64

base64 utility acts as a base64 decoder when passed the --decode (or -d) flag and as a base64 encoder
otherwise.  As a decoder it only accepts raw base64 input and as an encoder it does not produce the framing
lines.  base64 reads standard input or file if it is provided and writes to standard output.  Options
--wrap (or -w) and --ignore-garbage (or -i) are accepted for compatibility with GNU base64, but the latter
is unimplemented and silently ignored.

```text
Usage:
  ops -base64 [options] <string>
```

## Options

```
  -h, --help             Display this help message
  -e, --encode <string>  Encode a string to base64
  -d, --decode <string>  Decode a base64 string
```

## Examples

### Encoding: 

```bash
ops -base64 -e "OpenServerless is wonderful"
T3BlblNlcnZlcmxlc3MgaXMgd29uZGVyZnVs
```

### Decoding: 

```bash
$ ops -base64 -d "T3BlblNlcnZlcmxlc3MgaXMgd29uZGVyZnVs"
OpenServerless is wonderful
```