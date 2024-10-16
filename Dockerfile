FROM caddy:2.8.4-builder-alpine AS builder
COPY . .
RUN xcaddy build \
    --with github.com/kingreatwill/caddy-modules/markdown@latest=./markdown \
    --with github.com/kingreatwill/caddy-modules/tracing-sentry@latest=./tracing-sentry \
    --with github.com/kingreatwill/caddy-modules/admin-ui@latest=./admin-ui \
    --with github.com/caddyserver/forwardproxy@caddy2 \
    --with github.com/caddy-dns/dnspod@latest \
    --with github.com/caddyserver/nginx-adapter@latest

FROM caddy:2.8.4-alpine
LABEL maintainer "wcoder <350840291@qq.com>"
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
# validate install
# RUN /usr/bin/caddy -version
# RUN /usr/bin/caddy -plugins
# --with github.com/kingreatwill/caddy-modules/markdown@v1.0.3 \
# --with github.com/kingreatwill/caddy-modules/tracing-sentry@v1.0.3 \

# docker build --no-cache -t caddy-markdown:v0.0.1 .
# `caddy run|start --config nginx.conf --adapter nginx`
