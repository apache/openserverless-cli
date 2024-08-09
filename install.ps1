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

$SUFFIX = "_windows_amd64"
$EXT = ".zip"
$CMD = "ops.exe"

if (-not $IsWindows) {
    $OS = (uname -s)
    $ARCH = (uname -m)
    $CMD = "ops"
    switch ("$OS-$ARCH") {
        "Linux-x86_64" {
            $SUFFIX = "_linux_amd64"
            $EXT = ".tar.gz"
        }
        "Linux-arm64" {
            $SUFFIX = "_linux_arm64"
            $EXT = ".tar.gz"
        }
        "Darwin-x86_64" {
            $SUFFIX = "_darwin_amd64"
            $EXT = ".tar.gz"
        }
        "Darwin-arm64" {
            $SUFFIX = "_darwin_arm64"
            $EXT = ".tar.gz"
        }
        default {
            Write-Host "Unknown system - exiting"
            exit 1
        }
    }
} 

$OPSROOT = "https://raw.githubusercontent.com/apache/openserverless-task/0.1.0/opsroot.json"
$VERSION = (Invoke-RestMethod -Uri $OPSROOT).version
$FILE = "openserverless-cli_${VERSION}$SUFFIX$EXT"
$URL = "https://github.com/apache/openserverless-cli/releases/download/v$VERSION/$FILE"

# Create the directory if it doesn't exist
$BinPath = "$HOME/.local/bin"
if (-not (Test-Path -Path $BinPath)) {
    New-Item -ItemType Directory -Force -Path $BinPath | Out-Null
}

# Download the file
Invoke-WebRequest -Uri $URL -OutFile "$FILE"

# Unpack the file based on its extension
if ($EXT -eq ".zip") {
    Expand-Archive -Path "$FILE" -DestinationPath $BinPath -Force
} else {
    tar -xzvf "$FILE" -C $BinPath $CMD
}

# Verify installation
if (-not (Test-Path "$BinPath/$CMD*")) {
    Write-Host "Cannot install ops - download and unpack it in a folder in the path from here:"
    Write-Host $URL
    exit 1
}

# Check if the bin path is in the user's PATH
if (-not ($env:PATH -contains $BinPath)) {
    if($IsWindows) {
        $existingPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
            $newPath = "$BinPath;$existingPath"
            [System.Environment]::SetEnvironmentVariable("Path", $newPath, [System.EnvironmentVariableTarget]::User)
            Write-Host "Please restart your terminal to find ops in your path"
    } else { 
        Write-Host "$BinPath is not in the PATH - adding it"
        Add-Content -Path "$HOME\.bashrc" -Value "`nexport PATH=`"$BinPath`:`$PATH`""
        Add-Content -Path "$HOME\.zshrc" -Value "`nexport PATH=`"$BinPath`:`$PATH`""
        Write-Host "Please restart your terminal to find ops in your path"
    }
}

exit 0
