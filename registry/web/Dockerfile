# Stage 1: build
FROM oven/bun:1-alpine AS builder
WORKDIR /app
COPY package.json bun.lock* ./
RUN bun install --frozen-lockfile
COPY . .
ARG VITE_API_BASE_URL=https://registry.hivemind.rithul.dev
ARG VITE_REGISTRY_URL=https://registry.hivemind.rithul.dev/simple/
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
ENV VITE_REGISTRY_URL=$VITE_REGISTRY_URL
RUN bun run build

# Stage 2: serve with nginx
FROM nginx:alpine
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
