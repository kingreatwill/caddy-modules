FROM caddy:2.8.4-builder-alpine AS builder
COPY . .
# RUN apk add --no-cache gcc musl-dev
# RUN CGO_ENABLED=1 xcaddy build \
RUN xcaddy build \
    --with github.com/kingreatwill/caddy-modules/markdown@latest=./markdown \
    --with github.com/caddyserver/forwardproxy@caddy2 \
    --with github.com/caddy-dns/dnspod@latest \
    --with github.com/caddyserver/nginx-adapter@latest

FROM caddy:2.8.4-alpine
LABEL maintainer="wcoder <350840291@qq.com>"
COPY --from=builder /usr/bin/caddy /usr/bin/caddy


# # Minimum, at-least-to-install packages to run/build
# apk add --no-cache gcc musl-dev
# # Base meta-package to run/build (This includes gcc and musl-dev as well)
# apk add --no-cache build-base
# # Better to be installed packages for dev
# apk add --no-cache alpine-sdk build-base

# validate install
# RUN /usr/bin/caddy -version
# RUN /usr/bin/caddy -plugins
# --with github.com/kingreatwill/caddy-modules/markdown@v1.0.3 \
# --with github.com/kingreatwill/caddy-modules/tracing-sentry@v1.0.3 \
# --with github.com/kingreatwill/caddy-modules/tracing-sentry@latest=./tracing-sentry \
# --with github.com/kingreatwill/caddy-modules/admin-ui@latest=./admin-ui \
# docker build --no-cache -t caddy-markdown:v0.0.1 .
# `caddy run|start --config nginx.conf --adapter nginx`
