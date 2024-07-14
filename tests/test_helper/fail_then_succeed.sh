#!/bin/bash

file_path=".test_fail_then_succeed"

if [ ! -f "$file_path" ]; then
    touch "$file_path"
    exit 1
else
    rm "$file_path"
    exit 0
fi
