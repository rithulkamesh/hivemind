# Deployment Guide

## Architecture

The hivemind project deploys two services to a single EC2 instance:

- **Registry** (`registry.hivemind.rithul.dev`) — Go API + Node.js web app + Postgres, running via Docker Compose behind Caddy
- **Docs** (`hivemind.rithul.dev`) — Static Docusaurus site served by the same Caddy instance

## CI/CD Workflows

### `registry-deploy.yml` — Registry Deploy

Triggers on push to `main` touching `registry/**`, or manual dispatch.

| Job | Description |
|-----|-------------|
| `test-api` | Go tests with Postgres service container (skipped on force deploy) |
| `test-web` | Bun typecheck (skipped on force deploy) |
| `build-and-push` | Build arm64 Docker images, push to ECR with `:latest` + `:sha` tags |
| `deploy` | SSH via Instance Connect, pull images, rolling restart (web → api) |

Manual dispatch supports a `force` boolean to skip tests.

### `docs-deploy.yml` — Docs Deploy

Triggers on push to `main` touching `website/**`, or manual dispatch.

Single job: build Docusaurus, rsync `website/build/` to EC2 `/opt/hivemind-docs/`, verify with curl.

### `docs.yml` — Docs Versioning

Triggers on release published. Creates a versioned docs snapshot via `docusaurus docs:version` and opens a PR to merge it into main.

## GitHub Configuration

### Repository Variables (`vars.*`)

| Variable | Description | Example |
|----------|-------------|---------|
| `AWS_ACCOUNT_ID` | AWS account ID | `935761084809` |
| `REGISTRY_EC2_HOST` | EC2 public IP/hostname | `44.219.173.39` |
| `REGISTRY_EC2_INSTANCE_ID` | EC2 instance ID | `i-0ad1fdc8775243aee` |

### IAM Role (OIDC)

All workflows authenticate to AWS via OIDC federation — no static AWS keys are stored in GitHub.

- **Role**: `arn:aws:iam::<account-id>:role/hivemind-registry-deploy`
- **Trust policy**: Allows `sts:AssumeRoleWithWebIdentity` from the GitHub OIDC provider for this repository
- **Permissions**: ECR push/pull, EC2 Instance Connect `SendSSHPublicKey`, S3 (for package storage)

### Environment Secrets (on EC2)

The file `/opt/hivemind-registry/.env.prod` contains production secrets:

```
ECR_REGISTRY=<account-id>.dkr.ecr.us-east-1.amazonaws.com
POSTGRES_USER=registry
POSTGRES_PASSWORD=<password>
POSTGRES_DB=registry
JWT_SECRET=<secret>
INTERNAL_SECRET=<secret>
BASE_URL=https://registry.hivemind.rithul.dev
AWS_REGION=us-east-1
BUCKET_NAME=<s3-bucket>
GITHUB_CLIENT_ID=<id>
GITHUB_CLIENT_SECRET=<secret>
GOOGLE_CLIENT_ID=<id>
GOOGLE_CLIENT_SECRET=<secret>
```

## EC2 Setup (One-Time)

### Prerequisites

- Ubuntu 24.04 ARM64 (`t4g.small`)
- Docker + Docker Compose installed
- AWS CLI installed (for ECR login from the instance itself)
- IAM instance profile with ECR pull access

### Registry Setup

```bash
sudo mkdir -p /opt/hivemind-registry
sudo chown ubuntu:ubuntu /opt/hivemind-registry
cd /opt/hivemind-registry

# Copy docker-compose.prod.yml and Caddyfile from the repo
# Create .env.prod with all required secrets (see above)

# Start the stack
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d
```

### Docs Setup

```bash
sudo mkdir -p /opt/hivemind-docs
sudo chown ubuntu:ubuntu /opt/hivemind-docs
```

The docs-deploy workflow will rsync the built Docusaurus site to `/opt/hivemind-docs/`. Caddy serves it via the `hivemind.rithul.dev` server block in the Caddyfile.

### Caddyfile

The Caddyfile at `/opt/hivemind-registry/Caddyfile` serves both domains:

- `registry.hivemind.rithul.dev` → reverse proxy to API (`:8080`) and Web (`:3000`)
- `hivemind.rithul.dev` → static files from `/srv/docs` (host bind-mount from `/opt/hivemind-docs`)

After updating the Caddyfile, reload Caddy:

```bash
cd /opt/hivemind-registry
docker compose -f docker-compose.prod.yml --env-file .env.prod exec caddy caddy reload --config /etc/caddy/Caddyfile
```

### Database Migrations

Migrations are embedded in the API binary and run automatically on startup. If you need to run them manually:

```bash
cd /opt/hivemind-registry
docker compose -f docker-compose.prod.yml --env-file .env.prod exec api /registry migrate
```

## Manual Deploy

### Registry

Trigger via GitHub Actions → "Registry Deploy" → Run workflow (optionally check "Force deploy").

Or SSH to EC2 and pull/restart manually:

```bash
cd /opt/hivemind-registry
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin <ecr-registry>
docker compose -f docker-compose.prod.yml --env-file .env.prod pull api web
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --no-deps web
docker compose -f docker-compose.prod.yml --env-file .env.prod up -d --no-deps api
docker image prune -f
```

### Docs

Trigger via GitHub Actions → "Docs Deploy" → Run workflow.

Or build locally and rsync:

```bash
cd website && npm ci && npm run build
rsync -azP --delete website/build/ ubuntu@<ec2-host>:/opt/hivemind-docs/
```
