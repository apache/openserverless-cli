#!/bin/sh

if jq --version >/dev/null
then 
    LATEST=https://api.github.com/repos/apache/openserverless-cli/releases/latest
    VERSION=$(curl -s  "$LATEST" | jq -r '.name | .[1:]')
    curl -sL bit.ly/get-ops | VERSION=$VERSION bash
else 
    echo "Sorry for the annoyance, but I need jq installed to run..."
fi