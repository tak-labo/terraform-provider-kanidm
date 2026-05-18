#!/usr/bin/env bash
# Initializes Kanidm for acceptance testing.
# Recovers admin accounts, creates a service account, and writes token to .env.test.

set -euo pipefail

KANIDM_URL="https://localhost:8443"
ENV_FILE=".env.test"
CONTAINER="kanidm-test"

kanidm_auth_token() {
  local username="$1"
  local password="$2"

  # Step 1: init
  local step1
  step1=$(curl -fsk -X POST "$KANIDM_URL/v1/auth" \
    -H "Content-Type: application/json" \
    -D /tmp/kanidm_h1 \
    -d "{\"step\":{\"init\":\"$username\"}}" 2>/dev/null)

  local session_id cookie
  session_id=$(echo "$step1" | python3 -c "import sys,json; print(json.load(sys.stdin).get('sessionid',''))" 2>/dev/null || true)
  cookie=$(grep -i 'set-cookie.*auth-session' /tmp/kanidm_h1 2>/dev/null | sed 's/set-cookie: //i' | awk '{print $1}' | tr -d '\r' || true)

  if [ -z "$session_id" ]; then return 1; fi

  # Step 2: begin password
  curl -fsk -X POST "$KANIDM_URL/v1/auth" \
    -H "Content-Type: application/json" \
    -H "Cookie: $cookie" \
    -D /tmp/kanidm_h2 \
    -d "{\"sessionid\":\"$session_id\",\"step\":{\"begin\":\"password\"}}" > /dev/null 2>/dev/null || true

  local cookie2
  cookie2=$(grep -i 'set-cookie.*auth-session' /tmp/kanidm_h2 2>/dev/null | sed 's/set-cookie: //i' | awk '{print $1}' | tr -d '\r' || true)
  [ -z "$cookie2" ] && cookie2="$cookie"

  # Step 3: submit password
  local token
  token=$(curl -fsk -X POST "$KANIDM_URL/v1/auth" \
    -H "Content-Type: application/json" \
    -H "Cookie: $cookie2" \
    -d "{\"sessionid\":\"$session_id\",\"step\":{\"cred\":{\"password\":\"$password\"}}}" 2>/dev/null \
    | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('state',{}).get('success',''))" 2>/dev/null || true)

  echo "$token"
}

# Wait for Kanidm to be ready
echo "Waiting for Kanidm to be ready..."
for i in $(seq 1 30); do
  if curl -fsk "$KANIDM_URL/status" > /dev/null 2>&1; then
    echo "Kanidm is ready."
    break
  fi
  if [ "$i" -eq 30 ]; then
    echo "ERROR: Kanidm did not start in time."
    docker logs "$CONTAINER" 2>&1 | tail -10
    exit 1
  fi
  sleep 2
done

# Recover admin and idm_admin, capture passwords
echo "Recovering admin accounts..."
ADMIN_PASS=$(docker exec "$CONTAINER" kanidmd recover-account admin 2>/dev/null \
  | grep "new_password:" | awk '{print $NF}' | tr -d '"' || true)
IDM_ADMIN_PASS=$(docker exec "$CONTAINER" kanidmd recover-account idm_admin 2>/dev/null \
  | grep "new_password:" | awk '{print $NF}' | tr -d '"' || true)

if [ -z "$IDM_ADMIN_PASS" ]; then
  echo "ERROR: Could not recover idm_admin password"
  exit 1
fi

# Get idm_admin token
echo "Getting idm_admin token..."
IDM_ADMIN_TOKEN=$(kanidm_auth_token "idm_admin" "$IDM_ADMIN_PASS")

if [ -z "$IDM_ADMIN_TOKEN" ]; then
  echo "ERROR: Could not get idm_admin token"
  exit 1
fi

echo "Got idm_admin token."

# Create service account for tests
echo "Creating tf-test service account..."
curl -fsk -X POST "$KANIDM_URL/v1/service_account" \
  -H "Authorization: Bearer $IDM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"attrs":{"name":["tf-test"],"displayname":["Terraform Test Account"],"entry_managed_by":["idm_admins"]}}' \
  > /dev/null 2>&1 || echo "Service account may already exist."

# Grant required privileges
echo "Granting privileges to tf-test..."
for GROUP in idm_people_admins idm_group_admins idm_service_account_admins; do
  curl -fsk -X POST "$KANIDM_URL/v1/group/$GROUP/_attr/member" \
    -H "Authorization: Bearer $IDM_ADMIN_TOKEN" \
    -H "Content-Type: application/json" \
    -d '["tf-test@kanidm"]' > /dev/null 2>&1 || true
done

# Generate API token
echo "Generating API token..."
TOKEN=$(curl -fsk -X POST "$KANIDM_URL/v1/service_account/tf-test/_api_token" \
  -H "Authorization: Bearer $IDM_ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"label":"tf-test-token","expiry":null,"read_write":true}' 2>/dev/null \
  | tr -d '"' || true)

if [ -z "$TOKEN" ]; then
  echo "ERROR: Could not generate API token"
  exit 1
fi

# Write .env.test
cat > "$ENV_FILE" <<EOF
export KANIDM_URL=$KANIDM_URL
export KANIDM_TOKEN=$TOKEN
EOF

echo ""
echo "Setup complete!"
echo "Run: source $ENV_FILE && make testacc"
