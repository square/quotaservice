#!/bin/sh

cd .git/hooks
rm pre-commit
ln -s ../../bin/pre-commit

echo "Git hooks installed.  You're good to go."
