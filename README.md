# caddy-modules

https://caddyserver.com/docs/extending-caddy

## markdown
caddy markdown server

### debug

```
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
xcaddy build v2.7.5 --with github.com/kingreatwill/caddy-modules/markdown@v0.0.1=./markdown
chmod +x caddy
./caddy run
```

### template
1. simple
2. normal
3. Custom template files
```
markdown {
    template /markdown.tmpl
}

markdown {
    template normal
}
```