#!/bin/bash

# scripts/setup-template.sh
# Run after editing pkg/version/version.go

set -euo pipefail

VERSION_FILE="pkg/version/version.go"

if [ ! -f "$VERSION_FILE" ]; then
  echo "Error: $VERSION_FILE not found. Please create it first."
  exit 1
fi

echo "Reading values from $VERSION_FILE..."

# Robust extraction: find line with VarName = "value" and extract inside quotes
extract() {
  local var_name="$1"
  grep -E "${var_name}\s*=\s*\"" "$VERSION_FILE" | sed -E 's/.*=\s*"([^"]+)".*/\1/'
}

APP_NAME=$(extract "AppName")
APP_GO_PACKAGE=$(extract "GoPackage")
APP_SERVICE_NAME=$(extract "ServiceName")
APP_DB_SCHEMA=$(extract "DbSchemaName")
APP_NAME_KEBAB=$(extract "AppNameKebab")
APP_NAME_SNAKE=$(extract "AppNameSnake")
REPOSITORY=$(extract "Repository")

# Fallback checks
if [[ -z "$APP_NAME" || -z "$APP_GO_PACKAGE" || -z "$APP_SERVICE_NAME" || -z "$APP_DB_SCHEMA" || -z "$APP_NAME_KEBAB" || -z "$APP_NAME_SNAKE" || -z "$REPOSITORY" ]]; then
  echo "Error: Failed to extract one or more values from $VERSION_FILE. Check the file format."
  echo "Make sure each variable is on its own line like: VarName = \"value\""
  exit 1
fi

echo "Detected values:"
echo "  AppName:        $APP_NAME"
echo "  GoPackage:      $APP_GO_PACKAGE"
echo "  ServiceName:    $APP_SERVICE_NAME"
echo "  DbSchemaName:   $APP_DB_SCHEMA"
echo "  AppNameKebab:   $APP_NAME_KEBAB"
echo "  AppNameSnake:   $APP_NAME_SNAKE"
echo "  Repository:     $REPOSITORY"

read -p "Continue with these values? (y/n): " confirm
[[ "$confirm" =~ ^[yY]$ ]] || { echo "Aborted."; exit 0; }

# Step 1: Rename directories and files
if [[ -e "./cmd/template4YourProjectNameServer" ]]; then
  echo "Renaming cmd directories and files..."
  git mv ./cmd/template4YourProjectNameServer/template4YourProjectNameServer.go ./cmd/template4YourProjectNameServer/"$APP_NAME"Server.go
  git mv ./cmd/template4YourProjectNameServer/template4YourProjectNameServer_test.go ./cmd/template4YourProjectNameServer/"$APP_NAME"Server_test.go
  git mv ./cmd/template4YourProjectNameServer/template4YourProjectNameFront ./cmd/template4YourProjectNameServer/"$APP_NAME"Front
  git mv ./cmd/template4YourProjectNameServer ./cmd/"$APP_NAME"Server
else
  echo "./cmd/template4YourProjectNameServer was not found"
fi
if [[ -e "./pkg/template4gopackage" ]]; then
  echo "Renaming pkg directories and files..."
  git mv ./pkg/template4gopackage "./pkg/${APP_GO_PACKAGE}"
else
  echo "./pkg/template4gopackage was not found"
fi
if [[ -e "./api/proto/template_4_your_project_name" ]]; then
  echo "Renaming api/proto directories and files..."
  git mv ./api/proto/template_4_your_project_name/v1/template_4_your_project_name.proto  "./api/proto/template_4_your_project_name/v1/$APP_GO_PACKAGE.proto"
  git mv ./api/proto/template_4_your_project_name/ "./api/proto/$APP_GO_PACKAGE"
else
  echo "./api/proto/template_4_your_project_name was not found !"
fi

# Step 2: String replacements (safe delimiter |)
echo "Replacing strings in files..."
SED_INPLACE=(-i)

if sed --version >/dev/null 2>&1; then
  # GNU sed (Linux)
  SED_INPLACE=(-i)
else
  # BSD sed (macOS)
  SED_INPLACE=(-i '')
fi

find . -type f \
  \( -name '*.go' -o -name 'go.mod' -o -name '*.yaml' -o -name '*.yml' -o -name '*.proto' -o -name 'Dockerfile' -o -name 'Makefile' -o -name '*.md' -o -wholename 'scripts/buf_generate.sh' -o -name '*.env*' -o -name '*.sql' \) \
  ! -path './.git/*' ! -path './gen/*' \
  -print0 | xargs -0 sed "${SED_INPLACE[@]}" \
    -e "s|github.com/your-github-account/template-4-your-project-name|$REPOSITORY|g" \
    -e "s|template4gopackage|$APP_GO_PACKAGE|g" \
    -e "s|Template4ServiceName|$APP_SERVICE_NAME|g" \
    -e "s|template_4_your_project_name_db_schema|$APP_DB_SCHEMA|g" \
    -e "s|template-4-your-project-name|$APP_NAME_KEBAB|g" \
    -e "s|template_4_your_project_name|$APP_NAME_SNAKE|g" \
    -e "s|template4YourProjectName|$APP_NAME|g"

# Note: macOS sed uses -i '', Linux uses -i (no backup). This works on both.

# Step 4: Regenerate protobufs
echo "Regenerating protobuf code..."
if [ -f scripts/buf_generate.sh ]; then
  ./scripts/buf_generate.sh
elif command -v buf >/dev/null 2>&1; then
  buf generate
else
  echo "Warning: buf not found – please run 'buf generate' manually later."
fi

go mod tidy


echo ""
echo "Setup complete! Your project is now customized for '$APP_NAME'."
echo "Review changes, run tests, and push when ready."