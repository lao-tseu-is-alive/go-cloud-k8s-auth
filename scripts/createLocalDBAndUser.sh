#!/bin/bash
## createNewAppDBAndUser.sh
## version : 1.0.0
## script to create a postgresql user and database based on the given app name argument
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

DB_NAME="$APP_NAME_SNAKE"

echo "  will try to create database:     $DB_NAME"

read -p "Continue with these values? (y/n): " confirm
[[ "$confirm" =~ ^[yY]$ ]] || { echo "Aborted."; exit 0; }
cd /tmp || exit 1

#echo "## Converting App name Upper Case to underscore for database compatibility"
#DB_NAME=$(echo "APP_NAME_SNAKE" | sed --expression 's/\([A-Z]\)/_\L\1/g' --expression 's/^_//')
# generate a random password of 32 chars with chars selected in alphanumeric and some special chars
#DB_PASSWORD=`tr -dc '_+=()A-Z-a-z-0-9' < /dev/urandom | fold -w32 | head -n1`
#in this case i prefer to generate it with openssl, no user will enter this password manually
if DB_PASSWORD=$(openssl rand -base64 32); then
  echo "## Will try to create postgres user "
  echo "## username       : ${DB_NAME}"
  echo "## password       : ${DB_PASSWORD}"
  CREATE_USER="psql -c \"CREATE USER ${DB_NAME} WITH PASSWORD '${DB_PASSWORD}';\""
  echo "about to run : ${CREATE_USER}"
  su -c "${CREATE_USER}" postgres
  echo "## Will try to create database ${DB_NAME} with owner=${DB_NAME}"
  su -c "createdb -O ${DB_NAME} ${DB_NAME}" postgres
  # uncomment next line to add postgis extension to the db
  #su -c "psql -c 'CREATE EXTENSION postgis;' ${DB_NAME}" postgres
  su -c "psql -c 'CREATE EXTENSION unaccent;' ${DB_NAME}" postgres
  cd - || exit
  # https://www.freedesktop.org/software/systemd/man/systemd.service.html
  echo "## Will prepare a systemd unit conf file in current directory: ${APP_NAME}.conf"
  cat >"${APP_NAME}".conf <<EOS
[Service]
Environment="PORT=8080"
# a way to indicate which storage to use for now one of (memory|postgres)
Environment="DB_DRIVER=postgres"
Environment="DB_HOST=127.0.0.1"
Environment="DB_PORT=5432"
Environment="DB_NAME=${DB_NAME}"
Environment="DB_USER=${DB_NAME}"
Environment="DB_PASSWORD=${DB_PASSWORD}"
# in dev env it can be ok to disable SSL mode but in prod it is another story
# it depends on various factor. is your service (go) running in the same host as the db (localhost ?)
# if not, is the network between your server and your db trusted ?? read the doc and ask your security officer:
# https://www.postgresql.org/docs/11/libpq-ssl.html#LIBPQ-SSL-PROTECTION
Environment="DB_SSL_MODE=disable"
EOS
else
  echo "## 💥💥 ERROR: Failed to generate password with openssl rand [maybe try : sudo apt install rand]"
fi
