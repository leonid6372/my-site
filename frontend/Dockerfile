# syntax=docker/dockerfile:1
# Frontend image (plan §6.2): build with Node, serve with nginx.

FROM node:24-alpine AS build
WORKDIR /app
COPY package.json package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci
COPY . .
RUN npm run build

FROM nginx:1.29-alpine
# openssl creates the self-signed fallback certificate (docker/40-tls-certs.sh).
RUN apk add --no-cache openssl
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --chmod=0755 docker/40-tls-certs.sh /docker-entrypoint.d/40-tls-certs.sh
COPY --from=build /app/dist /usr/share/nginx/html
EXPOSE 80 443
HEALTHCHECK --interval=15s --timeout=3s --retries=3 CMD wget -q --no-check-certificate -O /dev/null https://127.0.0.1/ || exit 1
