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

### preview

https://note.wcoder.com/

![](preview.png)

### TODO
- 增加在线人数
- 文件访问次数
- 显示创建和修改时间
- 显示git提交信息和diff
- 排除文件夹
- 隐藏文件
- SEO
- markdown插件可配置