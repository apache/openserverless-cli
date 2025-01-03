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


OS="$(uname -s)"
ARCH="$(uname -m)"
CMD="ops"

case "$OS-$ARCH" in 
(Linux-x86_64)
   SUFFIX="_linux_amd64"
   EXT=".tar.gz"
;;
(Linux-aarch64)
   SUFFIX="_linux_arm64"
   EXT=".tar.gz"
;;
(Linux-arm64)   
   SUFFIX="_linux_arm64" 
   EXT=".tar.gz"
;;
(Darwin-x86_64) 
   SUFFIX="_darwin_amd64"
   EXT=".tar.gz" 
;;
(Darwin-arm64)  
  SUFFIX="_darwin_arm64"
  EXT=".tar.gz" 
;;
(MINGW64_NT-*)  
   SUFFIX="_windows_amd64" ; 
   EXT=".zip"
   CMD="ops.exe"
;;
(*) 
   echo "unknown system - exiting"
   exit 1 
;;
esac

OPSROOT="https://raw.githubusercontent.com/apache/openserverless-task/0.1.0/opsroot.json"
if test -z "$VERSION"
then VERSION="$(curl -sL $OPSROOT | sed -n 's/^.*"version": "\([^"]*\)",/\1/p')"
fi
FILE="openserverless-cli_${VERSION}$SUFFIX$EXT"
URL="https://github.com/apache/openserverless-cli/releases/download/v$VERSION/$FILE"

mkdir -p ~/.local/bin
curl -sL "$URL" -o "/tmp/$FILE"

if test "$EXT" == ".zip"
then 
   unzip -o -d ~/.local/bin "/tmp/$FILE" "$CMD"
else 
   tar xzvf "/tmp/$FILE" -C ~/.local/bin "$CMD"
fi


if ! test -e  ~/.local/bin/ops*
then echo "cannot install ops - download and unpack it in a folder in the path from here:"
     echo "$URL"
     exit 1
fi

if ! which ops | grep $HOME/.local/bin
then 
  echo "$HOME/.local/bin is not in the path - adding it"
  echo 'export PATH="$HOME/.local/bin:$PATH"' >>$HOME/.bashrc
  echo 'export PATH="$HOME/.local/bin:$PATH"' >>$HOME/.zshrc
  echo please restart your terminal to find ops in your path
fi

exit 0
