git checkout .
git pull

docker build --no-cache -t caddy-markdown:v0.0.1 .
docker stop caddy
docker rm caddy
docker run -d --cap-add=NET_ADMIN --restart=always --network host \
    -v /data/dockerv/caddy/srv:/srv \
    -v /data/dockerv/caddy/data:/data \
    -v /data/dockerv/caddy/log:/log \
    -v /data/dockerv/caddy/config:/config \
    -v /data/dockerv/caddy/Caddyfile:/etc/caddy/Caddyfile \
    --name caddy caddy-markdown:v0.0.1