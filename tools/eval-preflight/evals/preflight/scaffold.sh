#!/bin/sh
# A one-commit repository on main, as every vloop case's scaffold builds.
set -e
git init -q -b main
git config user.name preflight
git config user.email preflight@example.com
echo preflight > README.md
git add README.md
git commit -qm init
