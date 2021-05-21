#!/bin/sh
if [ $(uname) = "Darwin" ]; then
  "$(dirname "$0")"/darwin/main
else
  "$(dirname "$0")"/linux/main
fi
