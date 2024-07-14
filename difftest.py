#!/usr/bin/env python3
# Licensed to the Apache Software Foundation (ASF) under one
# or more contributor license agreements.  See the NOTICE file
# distributed with this work for additional information
# regarding copyright ownership.  The ASF licenses this file
# to you under the Apache License, Version 2.0 (the
# "License"); you may not use this file except in compliance
# with the License.  You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing,
# software distributed under the License is distributed on an
# "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
# KIND, either express or implied.  See the License for the
# specific language governing permissions and limitations
# under the License.


print("\n=== FAILING TESTS:")
import sys
tests = {}
base = "_difftest"
f = open(base, 'r')
lines = f.readlines()

fails=[]
gots = []
wants = []

i=0
while i < len(lines):
    line = lines[i]
    if(line.startswith("=== RUN")):
        if lines[i+1].startswith("--- FAIL:"):
            fails.append(i+1)
    i+=1

if len(sys.argv) == 1:
    n = 0
    for i in fails:
        print(n, lines[i], end='')
        n += 1
    print("=== use 'task utestdiff N=<n>' to see the diff")
    sys.exit(len(fails))

n = int(sys.argv[1])
k = fails[n]+2
got = []
want = []
while not lines[k].startswith("want:"):
    #print(lines[k])
    got.append(lines[k])
    k += 1
k += 1
while not (lines[k].startswith("FAIL") or lines[k].startswith("=== RUN")):
    #print(lines[k])
    want.append(lines[k])
    k += 1

import tempfile
import os

f1n = "%s.got"%base

f2n = "%s.want"%base

with open(f1n, "w") as f1:
    with open(f2n, "w") as f2:
        f1.writelines(got)
        f2.writelines(want)

os.system("diff %s %s" % (f1n, f2n))
