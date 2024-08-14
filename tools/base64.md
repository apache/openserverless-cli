`base64` utility acts as a base64 decoder when passed the `--decode` (or -d) flag and as a base64 encoder
otherwise. As a decoder it only accepts raw base64 input and as an encoder it does not produce the framing
lines.

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

### Encoding

```bash
ops -base64 -e "OpenServerless is wonderful"
```

This will output:

```text
T3BlblNlcnZlcmxlc3MgaXMgd29uZGVyZnVs
```

### Decoding

```bash
ops -base64 -d "T3BlblNlcnZlcmxlc3MgaXMgd29uZGVyZnVs"
```

This will output:

```text
OpenServerless is wonderful
```