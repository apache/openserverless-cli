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

Check if a value is valid according to the given constraints.
If -e is specified, the value is retrieved from the environment variable with the given name.

```text
Usage:
ops -validate [-e] [-m | -n | -r <regex>] <value> [<message>]
```

## Options

```
-e    Retrieve value from the environment variable with the given name.
-h    Print this help message.
-m    Check if the value is a valid email address.
-n    Check if the value is a number.
-r string Check if the value matches the given regular expression.
```

## Examples

### Validate with regexp

### Validate email

```
ops -validate -m example@gmail.com
```

```bash
ops -validate -r '^[a-z]+$' abc
```



