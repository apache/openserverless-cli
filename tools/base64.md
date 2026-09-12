<!--
  ~ Licensed to the Apache Software Foundation (ASF) under one
  ~ or more contributor license agreements.  See the NOTICE file
  ~ distributed with this work for additional information
  ~ regarding copyright ownership.  The ASF licenses this file
  ~ to you under the Apache License, Version 2.0 (the
  ~ "License"); you may not use this file except in compliance
  ~ with the License.  You may obtain a copy of the License at
  ~
  ~   http://www.apache.org/licenses/LICENSE-2.0
  ~
  ~ Unless required by applicable law or agreed to in writing,
  ~ software distributed under the License is distributed on an
  ~ "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  ~ KIND, either express or implied.  See the License for the
  ~ specific language governing permissions and limitations
  ~ under the License.
-->

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