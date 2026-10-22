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