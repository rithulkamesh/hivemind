# Registry dev — run from registry/
# Install: https://github.com/casey/just
# Usage: cd registry && just <recipe>
# Copy .env.example to .env and set DATABASE_URL (and OAuth vars if you want social login).

set dotenv-load

ECR_REGISTRY := env("ECR_REGISTRY", "935761084809.dkr.ecr.us-east-1.amazonaws.com")
EC2_HOST := env("EC2_HOST", "44.219.173.39")
EC2_INSTANCE_ID := env("EC2_INSTANCE_ID", "i-0ad1fdc8775243aee")

# Start Postgres + Mailhog + MinIO in Docker. Then run: just dev
deps:
    docker compose -f docker-compose.dev.yml up postgres mailhog minio createbuckets -d
    @echo "Waiting for postgres..."
    sleep 3
    @echo "Deps up (postgres, mailhog, minio). Run: just dev  (API + web in one terminal, Ctrl+C kills all)"

# Run API (air) + Web (Vite + Better Auth). Ctrl+C kills all. Requires .env with DATABASE_URL.
dev: deps
    #!/usr/bin/env bash
    set -e
    if [ -z "${DATABASE_URL:-}" ]; then echo "Error: DATABASE_URL not set. Copy .env.example to .env and set it."; exit 1; fi
    trap 'kill $(jobs -p) 2>/dev/null; exit 0' INT TERM
    (cd api && air) &
    # Wait for API to be ready before starting web
    echo "Waiting for API on :8080..."
    for i in $(seq 1 30); do
      if curl -sf http://localhost:8080/health >/dev/null 2>&1; then break; fi
      sleep 1
    done
    echo "API ready."
    (cd web && ( [ -d node_modules ] || bun install ) && exec bun run dev) &
    wait

# Run API only (requires Postgres). Uses .env.
api:
    cd api && air

# Run Web only (Vite + Better Auth). Proxies /api and /simple to localhost:8080.
web:
    cd web && bun run dev

# Run Go API migrations (packages, orgs, etc.). Requires Postgres. Uses .env.
db-migrate:
    cd api && MIGRATE_ONLY=1 go run ./cmd/registry

# Run Better Auth migrations (user, session, account, verification, jwks, twoFactor, etc.).
# Requires Postgres and .env with DATABASE_URL. Run from registry/ so .env is loaded.
auth-migrate:
    #!/usr/bin/env bash
    set -e
    cd "{{ justfile_directory() }}"
    [ -f .env ] && set -a && source .env && set +a
    if [ -z "${DATABASE_URL:-}" ]; then echo "Error: DATABASE_URL not set. Copy .env.example to .env in registry/."; exit 1; fi
    cd web && npx auth@latest migrate --config ./server/auth.ts --yes

# Full stack in Docker (postgres + api + web). No host watch.
dev-docker:
    docker compose -f docker-compose.dev.yml up

# Build test plugin wheel/sdist
test-plugin-build:
    cd test-plugin && pip install build && python -m build

# Install test plugin from local registry (registry must be up and package uploaded or seeded)
test-plugin-install:
    pip install --index-url http://localhost:3000/simple/ hivemind-plugin-demo

# --- Production Deploy ---

# Build and push API image to ECR (arm64)
build-api:
    docker buildx build --platform linux/arm64 -t {{ECR_REGISTRY}}/hivemind-registry-api:latest --push ./api

# Build and push Web image to ECR (arm64)
build-web:
    docker buildx build --platform linux/arm64 -t {{ECR_REGISTRY}}/hivemind-registry-web:latest --push ./web

# Build and push both images
build-all: build-api build-web

# SSH to EC2 via Instance Connect (generates ephemeral key, valid ~60s)
[private]
ssh-connect cmd:
    #!/usr/bin/env bash
    set -e
    ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q -y 2>/dev/null || true
    [ -f /tmp/ec2_deploy_key ] || ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q
    aws ec2-instance-connect send-ssh-public-key \
      --instance-id "{{EC2_INSTANCE_ID}}" \
      --instance-os-user ubuntu \
      --ssh-public-key file:///tmp/ec2_deploy_key.pub
    ssh -i /tmp/ec2_deploy_key -o StrictHostKeyChecking=no ubuntu@{{EC2_HOST}} "{{cmd}}"

# Deploy: pull latest images on EC2 and restart containers
deploy:
    #!/usr/bin/env bash
    set -e
    ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q 2>/dev/null || true
    aws ec2-instance-connect send-ssh-public-key \
      --instance-id "{{EC2_INSTANCE_ID}}" \
      --instance-os-user ubuntu \
      --ssh-public-key file:///tmp/ec2_deploy_key.pub
    ssh -i /tmp/ec2_deploy_key -o StrictHostKeyChecking=no ubuntu@{{EC2_HOST}} \
      "aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin {{ECR_REGISTRY}} && \
       cd /opt/hivemind-registry && \
       docker compose -f docker-compose.prod.yml --env-file .env.prod pull api web && \
       docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --no-deps api web && \
       docker image prune -f"

# Build, push, and deploy everything
ship: build-all deploy smoke

# Smoke tests against production
smoke:
    #!/usr/bin/env bash
    set -e
    echo "Running smoke tests against https://registry.hivemind.rithul.dev ..."
    curl -sf https://registry.hivemind.rithul.dev/api/v1/packages > /dev/null
    echo "  ✓ API packages endpoint"
    curl -sf https://registry.hivemind.rithul.dev/auth/jwks > /dev/null
    echo "  ✓ JWKS endpoint"
    curl -sf https://registry.hivemind.rithul.dev/ > /dev/null
    echo "  ✓ Frontend"
    curl -sf https://hivemind.rithul.dev/ > /dev/null
    echo "  ✓ Docs site"
    echo "All smoke tests passed!"

# View production logs (last 50 lines)
logs service="api":
    #!/usr/bin/env bash
    set -e
    ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q 2>/dev/null || true
    aws ec2-instance-connect send-ssh-public-key \
      --instance-id "{{EC2_INSTANCE_ID}}" \
      --instance-os-user ubuntu \
      --ssh-public-key file:///tmp/ec2_deploy_key.pub
    ssh -i /tmp/ec2_deploy_key -o StrictHostKeyChecking=no ubuntu@{{EC2_HOST}} \
      "docker logs hivemind-registry-{{service}}-1 --tail 50"

# Open SSH session to EC2
ssh:
    #!/usr/bin/env bash
    set -e
    ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q 2>/dev/null || true
    aws ec2-instance-connect send-ssh-public-key \
      --instance-id "{{EC2_INSTANCE_ID}}" \
      --instance-os-user ubuntu \
      --ssh-public-key file:///tmp/ec2_deploy_key.pub
    ssh -i /tmp/ec2_deploy_key -o StrictHostKeyChecking=no ubuntu@{{EC2_HOST}}

# Run production DB query
db-query query:
    #!/usr/bin/env bash
    set -e
    ssh-keygen -t ed25519 -f /tmp/ec2_deploy_key -N "" -q 2>/dev/null || true
    aws ec2-instance-connect send-ssh-public-key \
      --instance-id "{{EC2_INSTANCE_ID}}" \
      --instance-os-user ubuntu \
      --ssh-public-key file:///tmp/ec2_deploy_key.pub
    ssh -i /tmp/ec2_deploy_key -o StrictHostKeyChecking=no ubuntu@{{EC2_HOST}} \
      "docker exec hivemind-registry-postgres-1 psql -U registry -d hivemind_registry -c \"{{query}}\""
