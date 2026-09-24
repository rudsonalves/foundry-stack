#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
MOBILE_ENV="$REPO_ROOT/mobile/dev.env"
PORT=8080
CUSTOM_IP=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --ip) CUSTOM_IP="${2:-}"; shift 2 ;;
    --port) PORT="${2:-}"; shift 2 ;;
    *) echo "Erro: argumento desconhecido: $1" >&2; exit 1 ;;
  esac
done

get_lan_ip() {
  if command -v ip >/dev/null 2>&1; then
    ip -4 route get 1.1.1.1 2>/dev/null |
      awk '{for (i = 1; i <= NF; i++) if ($i == "src") {print $(i + 1); exit}}'
    return
  fi

  local interface
  interface="$(route get default 2>/dev/null | awk '/interface:/{print $2; exit}')"
  [[ -n "$interface" ]] && ipconfig getifaddr "$interface" 2>/dev/null || true
}

host_ip="${CUSTOM_IP:-$(get_lan_ip)}"
[[ -n "$host_ip" ]] || {
  echo "Erro: não foi possível detectar o IP local. Use --ip <endereço>." >&2
  exit 1
}
[[ -f "$MOBILE_ENV" ]] || {
  echo "Erro: mobile/dev.env não existe. Execute: make env-init" >&2
  exit 1
}

tmp="$(mktemp)"
awk -v url="http://${host_ip}:${PORT}" '
  BEGIN { replaced = 0 }
  /^BASE_URL=/ && !replaced { print "BASE_URL=" url; replaced = 1; next }
  { print }
  END { if (!replaced) print "BASE_URL=" url }
' "$MOBILE_ENV" > "$tmp"
mv "$tmp" "$MOBILE_ENV"
echo "Atualizado: mobile/dev.env (BASE_URL=http://${host_ip}:${PORT})"
