# Deployment Checklist

Before triggering a deployment, ensure the following steps are complete:

1. **Environment Variables**: Verify `.env.prod` on the EC2 instance in `/opt/registry/.env.prod` contains all essential secrets:
   - `JWT_SECRET`
   - `INTERNAL_SECRET`
   - `POSTGRES_USER`
   - `POSTGRES_PASSWORD`
   - `POSTGRES_DB`
   - `AWS_REGION`
   - `GITHUB_CLIENT_ID`
   - `GITHUB_CLIENT_SECRET`
   - `GOOGLE_CLIENT_ID`
   - `GOOGLE_CLIENT_SECRET`

2. **Caddyfile**: Ensure the `Caddyfile` is updated with the correct domain and is present in `/opt/registry/Caddyfile` on the host.

3. **Database Migrations**: Run migrations against the production database if the schema has changed. You can SSH in and run `docker compose -f docker-compose.prod.yml run api ...` or run locally pointing `DATABASE_URL` to production.

4. **Testing**: Run local smoke tests (`just smoke`) or unit tests to verify API endpoints are working properly before push.

5. **AWS Credentials**: Ensure GitHub Actions secrets are properly populated:
   - `AWS_ACCESS_KEY_ID`
   - `AWS_SECRET_ACCESS_KEY`
   - `AWS_REGION`
   - `ECR_REGISTRY` (e.g., `123456789012.dkr.ecr.us-east-1.amazonaws.com`)
   - `EC2_HOST` (e.g., `34.12.34.56`)
   - `EC2_USER` (e.g., `ubuntu`)
   - `EC2_KEY` (The raw private key content string)

Trigger deployment by pushing to the `main` branch.
