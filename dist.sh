#!/bin/sh
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
#
# Build one release archive for a single platform.
#
# The platform is selected with the standard Go environment variables:
#
#   GOOS=linux GOARCH=arm64 ./dist.sh
#
# and defaults to the host platform when they are unset. The result is
#
#   dist/openserverless-cli_<version>_<goos>_<goarch>.tar.gz   (.zip on windows)
#
# containing the `ops` binary next to LICENSE, DISCLAIMER and NOTICE.

set -e

cd "$(dirname "$0")"

PROJECT="openserverless-cli"
VERSION="$(cat version.txt)"

# Fall back to the host platform when GOOS/GOARCH are not set.
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
export GOOS GOARCH

# Windows gets a .exe binary in a zip; everyone else a plain binary in a tarball.
if test "$GOOS" = "windows"
then BIN="ops.exe"; EXT="zip"
else BIN="ops";     EXT="tar.gz"
fi

NAME="${PROJECT}_${VERSION}_${GOOS}_${GOARCH}"
STAGE="dist/$NAME"

rm -rf "$STAGE"
mkdir -p "$STAGE"

echo "building $NAME"

# Reproducible, statically linked, stripped: no cgo, no build paths embedded.
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "$STAGE/$BIN" ./cmd/ops/

# Every archive ships the ASF required legal files.
cp LICENSE DISCLAIMER NOTICE README.md "$STAGE"

# Archive from inside the staging dir so archive paths are not prefixed.
# List the members explicitly so they are stored without a "./" prefix:
# install.sh extracts a single member by name and that fails on "./ops".
FILES="LICENSE DISCLAIMER NOTICE README.md $BIN"
if test "$EXT" = "zip"
then (cd "$STAGE" && zip -q -X "../$NAME.zip" $FILES)
else (cd "$STAGE" && tar czf "../$NAME.tar.gz" $FILES)
fi

rm -rf "$STAGE"

ls -l "dist/$NAME.$EXT"
