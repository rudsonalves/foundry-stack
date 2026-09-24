#!/usr/bin/env bash
set -euo pipefail

# Inicializa os ambientes da API e do mobile sem substituir segredos existentes.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
CREDENTIALS_FILE="$REPO_ROOT/credentials.env"

random_hex_64() {
  openssl rand -hex 32
}

random_base64_32() {
  openssl rand -base64 32
}

extract_env_value() {
  local file="$1"
  local key="$2"

  [[ -f "$file" ]] || return 0
  awk -F= -v key="$key" '
    $1 == key {
      value = substr($0, index($0, "=") + 1)
      gsub(/^"|"$/, "", value)
      gsub(/^\047|\047$/, "", value)
      print value
      exit
    }
  ' "$file"
}

set_env_value() {
  local file="$1"
  local key="$2"
  local value="$3"
  local tmp

  tmp="$(mktemp)"
  awk -v key="$key" -v value="$value" '
    BEGIN { replaced = 0 }
    {
      if (!replaced && index($0, key "=") == 1) {
        print key "=" value
        replaced = 1
      } else {
        print
      }
    }
    END { if (!replaced) print key "=" value }
  ' "$file" > "$tmp"
  mv "$tmp" "$file"
}

is_placeholder() {
  local value="$1"
  [[ -z "$value" || "$value" == \<*\> || "$value" == "change-me" ]]
}

ensure_secret() {
  local file="$1"
  local key="$2"
  local generator="$3"
  local value

  value="$(extract_env_value "$file" "$key")"
  if is_placeholder "$value"; then
    set_env_value "$file" "$key" "$($generator)"
    echo "Gerado: $key em ${file#$REPO_ROOT/}"
  fi
}

sync_smtp_credentials() {
  local file="$1"
  local gmail_account gmail_password

  if [[ ! -f "$CREDENTIALS_FILE" ]]; then
    echo "Aviso: credentials.env não encontrado; credenciais SMTP não foram sincronizadas." >&2
    return
  fi

  gmail_account="$(extract_env_value "$CREDENTIALS_FILE" gmail_account)"
  gmail_password="$(extract_env_value "$CREDENTIALS_FILE" gmail_password)"

  if [[ -z "$gmail_account" || -z "$gmail_password" ]]; then
    echo "Erro: credentials.env deve definir gmail_account e gmail_password." >&2
    exit 1
  fi

  set_env_value "$file" SMTP_USERNAME "$gmail_account"
  set_env_value "$file" SMTP_PASSWORD "$gmail_password"
  set_env_value "$file" SMTP_FROM_ADDRESS "$gmail_account"
  echo "Sincronizado: credenciais SMTP em ${file#$REPO_ROOT/}"
}

configure_smtp() {
  local file="$1"

  set_env_value "$file" EMAIL_PROVIDER smtp
  set_env_value "$file" SMTP_HOST smtp.gmail.com
  set_env_value "$file" SMTP_PORT 587
  set_env_value "$file" SMTP_FROM_NAME FoundryStack
  set_env_value "$file" SMTP_TLS_MODE starttls
  sync_smtp_credentials "$file"
}

configure_email_verification_limits() {
  local file="$1"

  set_env_value "$file" EMAIL_VERIFICATION_TOKEN_TTL_SECONDS 86400
  set_env_value "$file" EMAIL_VERIFICATION_RESEND_COOLDOWN_SECONDS 60
  set_env_value "$file" EMAIL_VERIFICATION_RESEND_EMAIL_LIMIT_PER_HOUR 5
  set_env_value "$file" EMAIL_VERIFICATION_RESEND_IP_LIMIT_PER_HOUR 10
}

configure_password_reset_limits() {
  local file="$1"

  set_env_value "$file" PASSWORD_RESET_CODE_TTL_SECONDS 900
  set_env_value "$file" PASSWORD_RESET_TOKEN_TTL_SECONDS 900
  set_env_value "$file" PASSWORD_RESET_MAX_CONFIRM_ATTEMPTS 5
  set_env_value "$file" PASSWORD_RESET_RESEND_COOLDOWN_SECONDS 60
  set_env_value "$file" PASSWORD_RESET_EMAIL_LIMIT_PER_HOUR 5
  set_env_value "$file" PASSWORD_RESET_IP_LIMIT_PER_HOUR 10
  set_env_value "$file" PASSWORD_RESET_RETENTION_SECONDS 86400
}

configure_test_email() {
  local file="$REPO_ROOT/api/.env.test"
  local example="$REPO_ROOT/api/.env.test.example"

  if [[ ! -f "$file" ]]; then
    cp "$example" "$file"
    echo "Criado: ${file#$REPO_ROOT/}"
  fi

  set_env_value "$file" EMAIL_PROVIDER memory
  set_env_value "$file" SMTP_HOST ""
  set_env_value "$file" SMTP_PORT 587
  set_env_value "$file" SMTP_USERNAME ""
  set_env_value "$file" SMTP_PASSWORD ""
  set_env_value "$file" SMTP_FROM_ADDRESS no-reply@example.test
  set_env_value "$file" SMTP_FROM_NAME FoundryStack
  set_env_value "$file" SMTP_TLS_MODE starttls
  ensure_secret "$file" "JWT_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "EMAIL_VERIFICATION_CODE_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "PASSWORD_RESET_CODE_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "PASSWORD_RESET_TOKEN_SECRET_BASE64" random_base64_32
  configure_email_verification_limits "$file"
  configure_password_reset_limits "$file"
}

create_api_env() {
  local env="$1"
  local file="$REPO_ROOT/api/.env.$env"
  local example="$REPO_ROOT/api/.env.$env.example"
  local db_user db_password db_host db_port db_name database_url

  if [[ ! -f "$file" ]]; then
    cp "$example" "$file"
    echo "Criado: ${file#$REPO_ROOT/}"
  else
    echo "Preservado: ${file#$REPO_ROOT/}"
  fi

  ensure_secret "$file" "POSTGRES_PASSWORD" random_hex_64
  ensure_secret "$file" "JWT_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "APP_TOKEN" random_hex_64
  ensure_secret "$file" "EMAIL_VERIFICATION_CODE_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "PASSWORD_RESET_CODE_SECRET_BASE64" random_base64_32
  ensure_secret "$file" "PASSWORD_RESET_TOKEN_SECRET_BASE64" random_base64_32
  configure_smtp "$file"
  configure_email_verification_limits "$file"
  configure_password_reset_limits "$file"

  db_user="$(extract_env_value "$file" POSTGRES_USER)"
  db_password="$(extract_env_value "$file" POSTGRES_PASSWORD)"
  db_host="$(extract_env_value "$file" POSTGRES_BIND_ADDRESS)"
  db_port="$(extract_env_value "$file" POSTGRES_PORT)"
  db_name="$(extract_env_value "$file" POSTGRES_DB)"
  database_url="postgres://${db_user}:${db_password}@${db_host}:${db_port}/${db_name}?sslmode=disable"
  set_env_value "$file" DATABASE_URL "$database_url"
}

create_mobile_env() {
  local env="$1"
  local mode="$2"
  local base_url="$3"
  local api_file="$REPO_ROOT/api/.env.$env"
  local mobile_file="$REPO_ROOT/mobile/$env.env"
  local app_token

  app_token="$(extract_env_value "$api_file" APP_TOKEN)"
  if [[ ! -f "$mobile_file" ]]; then
    touch "$mobile_file"
    echo "Criado: ${mobile_file#$REPO_ROOT/}"
  else
    echo "Preservado: ${mobile_file#$REPO_ROOT/}"
  fi

  [[ -n "$(extract_env_value "$mobile_file" BASE_URL)" ]] || set_env_value "$mobile_file" BASE_URL "$base_url"
  set_env_value "$mobile_file" APP_MODE "$mode"
  set_env_value "$mobile_file" AUTH_CLIENT_ID "foundry-stack-mobile"
  [[ -n "$(extract_env_value "$mobile_file" CONNECT_TIMEOUT)" ]] || set_env_value "$mobile_file" CONNECT_TIMEOUT "30000"
  [[ -n "$(extract_env_value "$mobile_file" RECEIVE_TIMEOUT)" ]] || set_env_value "$mobile_file" RECEIVE_TIMEOUT "30000"
  set_env_value "$mobile_file" APP_ACCESS_TOKEN "$app_token"
}

command -v openssl >/dev/null 2>&1 || {
  echo "Erro: openssl não encontrado." >&2
  exit 1
}

create_api_env dev
create_api_env stag
create_api_env prod
configure_test_email

create_mobile_env dev dev "http://localhost:8080"
create_mobile_env stag staging "https://staging.example.com"
create_mobile_env prod prod "https://api.example.com"
