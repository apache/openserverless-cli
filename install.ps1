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

$InWindows = [System.Environment]::OSVersion.Platform -eq 'Win32NT'

if (-not $InWindows) {
    Write-Host "This script is only for Windows - exiting"
    exit 1
}

$SUFFIX = "_windows_amd64"
$EXT = ".zip"
$CMD = "ops.exe"

$OPSROOT = "https://raw.githubusercontent.com/apache/openserverless-task/0.1.0/opsroot.json"
$VERSION = (Invoke-RestMethod -Uri $OPSROOT).version
if (-not $VERSION) {
    Write-Host "Cannot determine version from $OPSROOT - exiting"
    exit 1
}
$FILE = "openserverless-cli_${VERSION}$SUFFIX$EXT"
$URL = "https://github.com/apache/openserverless-cli/releases/download/v$VERSION/$FILE"

# Create the directory if it doesn't exist
$BinPath = "$HOME\.local\bin"
if (-not (Test-Path -Path $BinPath)) {
    New-Item -ItemType Directory -Force -Path $BinPath | Out-Null
}

# Download the file to a temp location
$TmpFile = Join-Path ([System.IO.Path]::GetTempPath()) $FILE
try {
    Invoke-WebRequest -Uri $URL -OutFile "$TmpFile"
} catch {
    Write-Host "Cannot download ops from:"
    Write-Host $URL
    exit 1
}

# Unpack the file
Expand-Archive -Path "$TmpFile" -DestinationPath $BinPath -Force
Remove-Item -Path "$TmpFile" -Force -ErrorAction SilentlyContinue

# Verify installation
if (-not (Test-Path "$BinPath\$CMD")) {
    Write-Host "Cannot install ops - download and unpack it in a folder in the path from here:"
    Write-Host $URL
    exit 1
}

# Check if the bin path is in the user's PATH
if (-not (($env:PATH -split ';') -contains $BinPath)) {
    $existingPath = [System.Environment]::GetEnvironmentVariable("Path", [System.EnvironmentVariableTarget]::User)
    $newPath = "$BinPath;$existingPath"
    [System.Environment]::SetEnvironmentVariable("Path", $newPath, [System.EnvironmentVariableTarget]::User)
    Write-Host "Please restart your terminal to find ops in your path"
}

