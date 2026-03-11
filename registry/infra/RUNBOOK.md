# Registry infrastructure – first deploy runbook

## Prerequisites

- AWS CLI configured (`aws configure` or `AWS_PROFILE`)
- Terraform ≥ 1.5
- A Route53 hosted zone for your domain (e.g. `hivemind.rithul.dev`)
- An ACM certificate in **us-east-1** for `*.yourdomain` (e.g. `*.hivemind.rithul.dev`) — create in AWS Console if needed

## 1. Create Terraform state backend (one-time)

```bash
aws s3 mb s3://hivemind-terraform-state --region us-east-1
aws s3api put-bucket-versioning --bucket hivemind-terraform-state \
  --versioning-configuration Status=Enabled
aws dynamodb create-table --table-name hivemind-terraform-locks \
  --attribute-definitions AttributeName=LockID,AttributeType=S \
  --key-schema AttributeName=LockID,KeyType=HASH \
  --billing-mode PAY_PER_REQUEST \
  --region us-east-1
```

## 2. Configure variables

Copy the example and set values:

```bash
cp registry/infra/terraform.tfvars.example registry/infra/terraform.tfvars
# Edit terraform.tfvars: domain, registry_fqdn, packages_fqdn, ssh_allowed_cidrs, github_org_repo, etc.
```

If you don’t use the default domain, set `route53_zone_id` to your hosted zone ID, or leave it null to look up by `domain`.  
If your ACM cert is created manually, set `acm_certificate_arn` (must be in `us-east-1`).

## 3. Apply Terraform

```bash
cd registry/infra
terraform init
terraform apply
```

Note the outputs:

- **ec2_elastic_ip** → use for GitHub secret `REGISTRY_EC2_HOST`
- **ecr_registry_url** → base URL for API image (e.g. `123456789.dkr.ecr.us-east-1.amazonaws.com/hivemind-registry-api`)
- **deploy_private_key** (sensitive) → use for GitHub secret `REGISTRY_DEPLOY_KEY` (only if `create_key_pair = true`)
- **aws_account_id** → use for GitHub variable `AWS_ACCOUNT_ID`
- **github_oidc_role_arn** → used in workflows (role name: `hivemind-registry-deploy`)

## 4. GitHub configuration

1. **Repository variable** (Settings → Secrets and variables → Actions → Variables):
   - `AWS_ACCOUNT_ID` = Terraform output `aws_account_id`

2. **Repository secrets** (Settings → Secrets and variables → Actions → Secrets):
   - `REGISTRY_EC2_HOST` = Terraform output `ec2_elastic_ip`
   - `REGISTRY_DEPLOY_KEY` = Terraform output `deploy_private_key` (paste the full private key, including `-----BEGIN ... -----` and `-----END ... -----`)

## 5. EC2 instance setup

SSH to the instance (use the key from Terraform if you created it):

```bash
ssh -i /path/to/terraform-key.pem ec2-user@<ec2_elastic_ip>
```

Create the env file and ensure the app can start:

```bash
sudo mkdir -p /opt/hivemind-registry
sudo touch /opt/hivemind-registry/.env
sudo chown ec2-user:ec2-user /opt/hivemind-registry/.env
```

Edit `/opt/hivemind-registry/.env` with at least:

- `POSTGRES_PASSWORD` — used by Postgres and API
- `DATABASE_URL` — e.g. `postgres://hivemind:<POSTGRES_PASSWORD>@postgres:5432/hivemind_registry?sslmode=disable`
- `API_IMAGE` — full ECR image:tag, e.g. `123456789.dkr.ecr.us-east-1.amazonaws.com/hivemind-registry-api:latest`
- `S3_BUCKET`, `S3_REGION`, `JWT_SECRET`, and any other vars your API needs

Create the `web_dist` directory so Caddy can serve the frontend (GitHub Actions will rsync into it):

```bash
sudo mkdir -p /opt/hivemind-registry/web_dist
sudo chown -R ec2-user:ec2-user /opt/hivemind-registry
```

Start the stack (or let user-data do it; if you need to start manually):

```bash
cd /opt/hivemind-registry
docker compose up -d
```

## 6. Trigger first deploy

- Push to `main` (or run the **Registry API** and **Registry Web** workflows manually).  
- API workflow: builds image, pushes to ECR, SSHs to EC2 and runs `docker compose pull api && docker compose up -d --no-deps api`.  
- Web workflow: builds the React app, rsyncs to `EC2:/opt/hivemind-registry/web_dist/`, then runs `docker exec ... caddy reload`.

If the Caddy container name differs (e.g. different project name), update the `docker exec` in `.github/workflows/registry-web.yml` (e.g. `hivemind-registry-caddy-1` → your container name).

## 7. SES

- In AWS Console → SES: request **production access** so you can send to non-verified addresses.
- Verify the domain: DKIM (and SPF/DMARC) records are created by Terraform in Route53; wait until SES shows the domain as verified.

## 8. Migration to ECS (later)

When you outgrow the single EC2 (e.g. sustained high CPU/memory or need zero-downtime deploys):

1. Add Terraform modules for ECS (cluster, task definitions, ALB, target groups).
2. Move Postgres to RDS or Aurora Serverless if desired.
3. Change deploy workflows to update ECS services instead of SSH/docker compose.
4. Point Route53 A record for the registry to the ALB.
5. Terminate the EC2 instance.

No application code changes are required; only infra and CI/CD.
