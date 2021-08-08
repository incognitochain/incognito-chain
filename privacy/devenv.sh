#!/bin/sh
# add a local pre-commit hook that runs linters & tests for `privacy` code
# NOTE: only do this once. In case of issues / need removing this hook, you will need to edit your .git/hooks/pre-commit manually
if ! [ -x "$(command -v golangci-lint)" ]; then
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sudo sh -s -- -b /usr/local/bin v1.41.1
fi
echo "./privacy/.githooks/pre-commit.sh" >> ../.git/hooks/pre-commit
chmod +x ../.git/hooks/pre-commit