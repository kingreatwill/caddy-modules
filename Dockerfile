FROM caddy:2.7.5-builder-alpine AS builder

RUN xcaddy build \
    --with github.com/kingreatwill/caddy-modules/markdown

FROM caddy:2.7.5-alpine
LABEL maintainer "wcoder <350840291@qq.com>"
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
# validate install
# RUN /usr/bin/caddy -version
# RUN /usr/bin/caddy -plugins

# docker build -t caddy-markdown:v0.0.1 .