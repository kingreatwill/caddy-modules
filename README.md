---
Title: wcoder 
Summary: caddy markdown
Tags:
    - markdown
    - caddy
---

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
- [ ] 增加在线人数
- [ ] 文件访问次数
- [ ] 显示创建和修改时间
- [ ] 显示git提交信息和diff
- [ ] 排除文件夹
- [x] 隐藏文件(目前默认以.和_开头的文件不显示)
- [x] SEO
- [ ] markdown插件可配置
- [ ] sitemap

#### 显示创建和修改时间
```bash
# 获取 git 仓库中所有文件的最新修改时间
$ git ls-tree -r --name-only HEAD | while read filename; do
> echo "$(git log -1 --pretty=format:"%ad" -- $filename) $filename";
> done
# 获取 git 仓库中所有文件的最初创建时间
$ git ls-tree -r --name-only HEAD | while read filename; do
> echo "$(git log --pretty=format:"%ad" -- $filename | tail -1) $filename";
> done
```

> [How to retrieve the last modification date of all files in a Git repository](https://serverfault.com/questions/401437/how-to-retrieve-the-last-modification-date-of-all-files-in-a-git-repository/401450#401450)

> [Finding the date/time a file was first added to a Git repository](https://stackoverflow.com/questions/2390199/finding-the-date-time-a-file-was-first-added-to-a-git-repository/2390382#2390382)
