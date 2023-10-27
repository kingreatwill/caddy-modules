package markdown

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/caddyconfig/caddyfile"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

type Markdown struct {
	Template string `json:"template,omitempty"`

	MIMETypes []string `json:"mime_types,omitempty"`
}

func (Markdown) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers.markdown",
		New: func() caddy.Module { return new(Markdown) },
	}
}

// Provision sets up the module. #caddy.Provisioner
func (md *Markdown) Provision(ctx caddy.Context) error {
	// TODO: set up the module
	return nil
}

// Validate ensures md has a valid configuration. #caddy.Validator
// Validate should be a read-only function. It is run after the Provision() method.
func (md *Markdown) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (md *Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) error {
	m.w.Write([]byte(r.RemoteAddr))
	return next.ServeHTTP(w, r)
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Markdown)(nil)
	_ caddy.Validator             = (*Markdown)(nil)
	_ caddyhttp.MiddlewareHandler = (*Markdown)(nil)
	_ caddyfile.Unmarshaler       = (*Markdown)(nil)
)