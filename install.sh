#!/bin/sh

OS="$(uname -s)"
ARCH="$(uname -m)"
CMD="ops"

case "$OS-$ARCH" in 
(Linux-x86_64)  
   SUFFIX="_linux_amd64"
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
VERSION="$(curl -sL $OPSROOT | sed -n 's/^.*"version": "\([^"]*\)",/\1/p')"
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