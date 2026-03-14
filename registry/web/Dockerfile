# Stage 1: install dependencies + build React SPA
FROM oven/bun:1-alpine AS builder
WORKDIR /app
COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile
COPY . .
ARG VITE_API_BASE_URL=https://registry.hivemind.rithul.dev
ARG VITE_REGISTRY_URL=https://registry.hivemind.rithul.dev/simple/
ARG VITE_GOOGLE_CLIENT_ID=
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_REGISTRY_URL=$VITE_REGISTRY_URL
ENV VITE_GOOGLE_CLIENT_ID=$VITE_GOOGLE_CLIENT_ID
RUN bun run build

# Stage 2: production runtime (Express + Better Auth + static files)
FROM oven/bun:1-alpine
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile --production
COPY server/ ./server/
COPY --from=builder /app/dist ./dist
EXPOSE 3000
CMD ["bun", "run", "server/index.ts"]
