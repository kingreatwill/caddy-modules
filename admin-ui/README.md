
## web
cd web
npm install
npm run build

## How to use

Build caddy with this package

```bash
xcaddy build --with github.com/kingreatwill/caddy-modules/admin-ui@v1.0.3
```

Add a http config

```
{
        admin localhost:2019
}

:2018 {
    route {
        admin_ui
        reverse_proxy localhost:2019 {
            header_up Host localhost:2019
        }
    }
}
```
or 只允许GET,防止被修改
```
{
    debug
    admin localhost:2019
    order admin_ui before reverse_proxy
}

:2018 {
    admin_ui
    reverse_proxy localhost:2019 {
        method GET
        header_up Host localhost:2019
    }
}
```

## Feature

- Show Server Config
- Show Upstream
- Show PKI
- Load Server Config and Save Config to Server
- View Metrics from "/metrics"