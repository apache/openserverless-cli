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
