package search

import (
	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/httpcaddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

const DirectiveName = "search"

func init() {
	caddy.RegisterModule(Search{})
	httpcaddyfile.RegisterHandlerDirective(DirectiveName, parseCaddyfile)
}

// parseCaddyfile sets up the handler from Caddyfile tokens. Syntax:
//
//	search {
//	    root <name>
//	    endpoint /search
//	    regexp *.md
//	}
func parseCaddyfile(h httpcaddyfile.Helper) (caddyhttp.MiddlewareHandler, error) {
	sch := new(Search)

	for h.Next() {
		for h.NextBlock(0) {
			switch h.Val() {
			case "endpoint":
				if !h.Args(&sch.Endpoint) {
					return nil, h.ArgErr()
				}
			case "root":
				if !h.Args(&sch.Root) {
					return nil, h.ArgErr()
				}
			case "regexp":
				if !h.Args(&sch.Regexp) {
					return nil, h.ArgErr()
				}
			}
		}
	}
	return sch, nil
}
