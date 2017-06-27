#!/bin/sh

# Licensed under the Apache License, Version 2.0
# Details: https://raw.githubusercontent.com/square/quotaservice/master/LICENSE

cd .git/hooks
rm pre-commit
ln -s ../../bin/pre-commit

echo "Git hooks installed.  You're good to go."
