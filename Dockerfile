FROM caddy:2.7.5-builder-alpine AS builder
COPY . .
RUN xcaddy build \
    --with github.com/kingreatwill/caddy-modules/markdown@latest=./markdown \
    --with github.com/caddyserver/forwardproxy@caddy2 \
    --with github.com/caddy-dns/dnspod@latest

FROM caddy:2.7.5-alpine
LABEL maintainer "wcoder <350840291@qq.com>"
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
# validate install
# RUN /usr/bin/caddy -version
# RUN /usr/bin/caddy -plugins

# docker build --no-cache -t caddy-markdown:v0.0.1 .
