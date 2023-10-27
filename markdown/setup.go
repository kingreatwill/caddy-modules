package markdown

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Markdown{})
	httpcaddyfile.RegisterHandlerDirective("markdown", parseCaddyfile)
}

// parseCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//     markdown [<matcher>] {
//         template <name>
//     }
//
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	md := new(Markdown)

	for h.Next() {
		for h.NextBlock(0) {
			switch h.Val() {
			case "template":
				if !h.Args(&md.Template) {
					return nil, h.ArgErr()
				}
			}
		}
	}
	return md, nil
}