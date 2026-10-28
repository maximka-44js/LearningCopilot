#!/usr/bin/env bash
set -euo pipefail

ENV_FILE="${1:-}"
PROFILE="${2:-unknown}"

if [[ -z "$ENV_FILE" ]]; then
  echo "[preflight] ERROR: env file path is required" >&2
  exit 1
fi

if [[ ! -f "$ENV_FILE" ]]; then
  echo "[preflight] ERROR: env file not found: $ENV_FILE" >&2
  exit 1
fi

declare -A env_map

while IFS= read -r line || [[ -n "$line" ]]; do
  [[ -z "$line" ]] && continue
  [[ "$line" =~ ^[[:space:]]*# ]] && continue
  [[ "$line" != *=* ]] && continue

  key="${line%%=*}"
  value="${line#*=}"

  key="${key#${key%%[![:space:]]*}}"
  key="${key%${key##*[![:space:]]}}"
  value="${value#${value%%[![:space:]]*}}"
  value="${value%${value##*[![:space:]]}}"

  if [[ "$value" =~ ^".*"$ ]]; then
    value="${value:1:-1}"
  fi

  env_map["$key"]="$value"
done < "$ENV_FILE"

required_vars=(
  POSTGRES_DB
  POSTGRES_USER
  POSTGRES_PASSWORD
  PG_URL
  KEYCLOAK_ADMIN
  KEYCLOAK_ADMIN_PASSWORD
  KEYCLOAK_DB_NAME
  KEYCLOAK_DB_USER
  KEYCLOAK_DB_PASSWORD
  S3_REGION_NAME
  S3_ACCESS_KEY
  S3_SECRET_KEY
  S3_BUCKET_NAME
  JWT_ISSUER
  JWT_CLIENT_ID
  FRONTEND_AUTH_SECRET
  FRONTEND_KEYCLOAK_CLIENT_ID
  FRONTEND_KEYCLOAK_CLIENT_SECRET
  FRONTEND_KEYCLOAK_ISSUER
)

if [[ "$PROFILE" == "prod" ]]; then
  required_vars=(PUBLIC_DOMAIN "${required_vars[@]}")
fi

bad_tokens=(
  "replace-with"
  "change-me"
  "your-dev-key-here"
  "your_auth_secret_here"
)

errors=()

for var_name in "${required_vars[@]}"; do
  value="${env_map[$var_name]:-}"
  if [[ -z "$value" ]]; then
    errors+=("$var_name is empty")
    continue
  fi

  lower_value="$(printf '%s' "$value" | tr '[:upper:]' '[:lower:]')"
  for token in "${bad_tokens[@]}"; do
    if [[ "$lower_value" == *"$token"* ]]; then
      errors+=("$var_name contains placeholder token '$token'")
      break
    fi
  done

done

if [[ -n "${env_map[FRONTEND_AUTH_SECRET]:-}" && ${#env_map[FRONTEND_AUTH_SECRET]} -lt 32 ]]; then
  errors+=("FRONTEND_AUTH_SECRET must contain at least 32 characters")
fi

if [[ -n "${env_map[FRONTEND_KEYCLOAK_CLIENT_SECRET]:-}" && ${#env_map[FRONTEND_KEYCLOAK_CLIENT_SECRET]} -lt 16 ]]; then
  errors+=("FRONTEND_KEYCLOAK_CLIENT_SECRET must contain at least 16 characters")
fi

pg_url="${env_map[PG_URL]:-}"
if [[ -n "$pg_url" && ! "$pg_url" =~ ^postgres(ql)?:// ]]; then
  errors+=("PG_URL must use the postgres:// or postgresql:// scheme")
fi

if [[ "$PROFILE" == "prod" ]]; then
  public_scheme="${env_map[PUBLIC_SCHEME]:-}"
  public_domain="${env_map[PUBLIC_DOMAIN]:-}"
  swagger_enabled="${env_map[SWAGGER_ENABLED]:-false}"

  if [[ "$public_scheme" != "https" ]]; then
    errors+=("PUBLIC_SCHEME must be https for production")
  fi
  if [[ "$public_domain" == *localhost* || "$public_domain" == *127.0.0.1* || "$public_domain" == *example.com* ]]; then
    errors+=("PUBLIC_DOMAIN must be a real production domain")
  fi
  if [[ "$swagger_enabled" != "false" ]]; then
    errors+=("SWAGGER_ENABLED must be false for production")
  fi
fi

if (( ${#errors[@]} > 0 )); then
  echo "[preflight] ERROR: invalid env for profile '$PROFILE' ($ENV_FILE)" >&2
  for err in "${errors[@]}"; do
    echo " - $err" >&2
  done
  exit 1
fi

echo "[preflight] OK: profile '$PROFILE' env looks valid ($ENV_FILE)"