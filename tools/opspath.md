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
