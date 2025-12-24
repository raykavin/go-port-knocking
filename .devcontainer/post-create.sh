#!/bin/sh
set -e

echo "> Configuring Git"

# Initialize git repo if it does not exist
if [ ! -d ".git" ]; then
  echo "Initializing Git repository (branch: $GIT_INIT_DEFAULT_BRANCH)"
  git init --initial-branch="$GIT_INIT_DEFAULT_BRANCH"
else
  echo "Git repository already exists, skipping init"
fi

# Configure safe directory and user
git config --global --add safe.directory "/workspaces/app"
git config --global user.name "$DEVELOPER_NAME"
git config --global user.email "$DEVELOPER_EMAIL"

# Configure remote upstream
if git remote | grep -q "^origin$"; then
  CURRENT_URL="$(git remote get-url origin)"
  if [ "$CURRENT_URL" != "$GIT_REPO_ADDRESS" ]; then
    echo "Updating origin remote URL"
    git remote set-url origin "$GIT_REPO_ADDRESS"
  else
    echo "Origin remote already configured correctly"
  fi
else
  echo "Adding origin remote"
  git remote add origin "$GIT_REPO_ADDRESS"
fi

# Ensure branch tracks origin
if git show-ref --verify --quiet "refs/heads/$GIT_INIT_DEFAULT_BRANCH"; then
  git branch --set-upstream-to="origin/$GIT_INIT_DEFAULT_BRANCH" "$GIT_INIT_DEFAULT_BRANCH" 2>/dev/null || true
fi

echo "> Adjusting Go directories permissions"
sudo chmod -R 777 /go/bin /go/pkg

echo "> Ensuring GOPATH is in PATH"
export PATH="$PATH:/go/bin"

echo "> Initializing Go modules"
if [ ! -f go.mod ]; then
  echo "Creating Go Module with project name: $PROJECT_NAME"
  go mod init "$PROJECT_NAME"
else
  echo "go.mod already exists, skipping module init"
fi

go mod tidy

echo "> Installing Go development tools"

go install golang.org/x/tools/gopls@latest
go install golang.org/x/tools/cmd/goimports@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

echo "> Verifying installed tools"
gopls version
dlv version
golangci-lint version

echo "> Running initial lint"
golangci-lint run || true

echo "Post-create completed successfully"
