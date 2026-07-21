#!/usr/bin/env bash

set -euo pipefail

API_URL="${API_URL:-http://127.0.0.1:8080}"

for command in curl jq; do
  if ! command -v "${command}" >/dev/null 2>&1; then
    echo "${command} is required" >&2
    exit 1
  fi
done

workdir="$(mktemp -d)"
trap 'rm -rf "${workdir}"' EXIT

create_headers="${workdir}/create.headers"
create_body="${workdir}/create.json"
create_status="$(curl --silent --show-error \
  --dump-header "${create_headers}" \
  --output "${create_body}" \
  --write-out '%{http_code}' \
  --request POST \
  "${API_URL}/api/v1/chats")"
[[ "${create_status}" == "201" ]]
grep -qi '^X-Request-ID:' "${create_headers}"
jq -e '.code == "OK" and .message == "success" and (.data.id > 0)' "${create_body}" >/dev/null
chat_id="$(jq -r '.data.id' "${create_body}")"

message_body="${workdir}/message.json"
message_status="$(curl --silent --show-error \
  --output "${message_body}" \
  --write-out '%{http_code}' \
  --header 'Content-Type: application/json' \
  --request POST \
  --data '{"content":"帮我查看当前网关路由"}' \
  "${API_URL}/api/v1/chats/${chat_id}/messages")"
[[ "${message_status}" == "201" ]]
jq -e --argjson chat_id "${chat_id}" \
  '.code == "OK" and .data.chat_id == $chat_id and .data.role == "user"' \
  "${message_body}" >/dev/null

chat_body="${workdir}/chat.json"
curl --silent --show-error "${API_URL}/api/v1/chats/${chat_id}" >"${chat_body}"
jq -e --argjson chat_id "${chat_id}" '.code == "OK" and .data.id == $chat_id' "${chat_body}" >/dev/null

messages_body="${workdir}/messages.json"
curl --silent --show-error \
  "${API_URL}/api/v1/chats/${chat_id}/messages?after_id=0&limit=50" >"${messages_body}"
jq -e --argjson chat_id "${chat_id}" \
  '.code == "OK" and (.data.items | length) == 1 and .data.items[0].chat_id == $chat_id' \
  "${messages_body}" >/dev/null

jq . "${create_body}"
jq . "${message_body}"
jq . "${chat_body}"
jq . "${messages_body}"
