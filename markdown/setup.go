package markdown

import (
	"mime"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func init() {
	caddy.RegisterModule(Markdown{})
	httpcaddyfile.RegisterHandlerDirective("markdown", parseCaddyfile)

	mime.AddExtensionType(".py", "text/plain")
	mime.AddExtensionType(".c", "text/plain")
	mime.AddExtensionType(".h", "text/plain")
	// mime.AddExtensionType(".go", "text/plain")
	mime.AddExtensionType(".cs", "text/plain")
	mime.AddExtensionType(".mod", "text/plain")
	mime.AddExtensionType(".sum", "text/plain")
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
			case "root":
				if !h.Args(&md.Root) {
					return nil, h.ArgErr()
				}
			case "hide":
				md.Hide = h.RemainingArgs()
				if len(md.Hide) == 0 {
					return nil, h.ArgErr()
				}
			case "index":
				md.IndexNames = h.RemainingArgs()
				if len(md.IndexNames) == 0 {
					return nil, h.ArgErr()
				}
			}
		}
	}
	return md, nil
}