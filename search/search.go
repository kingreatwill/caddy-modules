package search

import (
	"net/http"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"go.uber.org/zap"
)

type Search struct {
	Root     string `json:"root,omitempty"`
	Endpoint string `json:"endpoint,omitempty"` // default: /search
	Regexp   string `json:"regexp,omitempty"`
	logger   *zap.Logger
	watch    *NotifyFile
}

func (Search) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID:  "http.handlers." + DirectiveName,
		New: func() caddy.Module { return new(Search) },
	}
}

// Provision sets up the module. #caddy.Provisioner
func (sch *Search) Provision(ctx caddy.Context) error {
	sch.logger = ctx.Logger(sch)

	if sch.Root == "" {
		sch.Root = "{http.vars.root}"
	}
	if sch.Endpoint == "" {
		sch.Endpoint = "/search"
	}
	if sch.Regexp == "" {
		sch.Regexp = "*"
	}
	// 监听文件变化
	repl := ctx.Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	root := repl.ReplaceAll(sch.Root, ".")
	sch.watch = NewNotifyFile(sch.logger, sch.IndexDoc)
	if err := sch.watch.WatchDir(root); err != nil {
		return err
	}
	return nil
}

// Validate ensures md has a valid configuration. #caddy.Validator
// Validate should be a read-only function. It is run after the Provision() method.
func (sch *Search) Validate() error {
	return nil
}

// ServeHTTP implements caddyhttp.MiddlewareHandler.
func (sch *Search) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {
	return nil
}

// IndexDoc 索引文件
func (sch *Search) IndexDoc(path string, remove bool) {

}
