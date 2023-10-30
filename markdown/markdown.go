package markdown

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/kingreatwill/caddy-modules/markdown/convert"
	"github.com/kingreatwill/caddy-modules/markdown/template"
	"go.uber.org/zap"
)

type Markdown struct {
	Template  string   `json:"template,omitempty"`
	MIMETypes []string `json:"mime_types,omitempty"`
	engine    *convert.MarkdownConvert
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (Markdown) CaddyModule() caddy.ModuleInfo {
	return caddy.ModuleInfo{
		ID: "http.handlers.markdown",
		New: func() caddy.Module {
			return &Markdown{
				engine: convert.New(),
			}
		},
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
	_path := r.URL.Path
	caddy.Log().Info("ServeHTTP3:", zap.String("path", r.URL.Path))
	// if !strings.HasSuffix(_path, ".md") && !strings.HasSuffix(_path, ".markdown") {
	// 	return next.ServeHTTP(w, r)
	// }

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	shouldBuf := func(status int, header http.Header) bool {
		if strings.HasSuffix(_path, ".md") || strings.HasSuffix(_path, ".markdown") {
			return true
		}
		ct := header.Get("Content-Type")
		if strings.Contains(ct, "text/markdown") {
			return true
		}
		return false
	}
	rec := caddyhttp.NewResponseRecorder(w, buf, shouldBuf)
	err := next.ServeHTTP(rec, r)
	if err != nil {
		return err
	}
	if !rec.Buffered() {
		return nil
	}

	inputStr := buf.String()
	// template
	tmpl, ok := template.Templates[md.Template]
	if !ok {
		// if not a built-in template, try as resource file
		buf.Reset()
		fs := http.Dir(".")
		file, err := fs.Open(md.Template)
		if err == nil {
			defer file.Close()
			io.Copy(buf, file)
		}
		if buf.Len() > 0 {
			tmpl = buf.String()
		} else {
			tmpl = "{{.MdHtml}}"
		}
	}
	// render markdown
	html, err := md.renderMarkdown(r.Context(), inputStr, tmpl)
	if err != nil {
		return caddyhttp.Error(http.StatusInternalServerError, err)
	}

	buf.Reset()
	buf.WriteString(html)
	rec.Header().Set("Content-Type", "text/html; charset=utf-8")
	rec.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	rec.Header().Del("Accept-Ranges") // we don't know ranges for dynamically-created content
	rec.Header().Del("Last-Modified") // useless for dynamic content since it's always changing
	// we don't know a way to quickly generate etag for dynamic content,
	// and weak etags still cause browsers to rely on it even after a
	// refresh, so disable them until we find a better way to do this
	rec.Header().Del("Etag")
	return rec.WriteResponse()
}

func (md *Markdown) renderMarkdown(ctx context.Context, inputStr, tmplStr string) (string, error) {
	// TODO: 这里使用哪些markdown插件也是可以配置的
	data, err := md.engine.Convert(inputStr)
	if err != nil {
		return "", err
	}
	if data.Title == "" {
		orignalRequest := ctx.Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
		data.Title = path.Base(orignalRequest.URL.Path)
	}
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	// 解析模板
	err = template.Execute(buf, tmplStr, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Markdown)(nil)
	_ caddy.Validator             = (*Markdown)(nil)
	_ caddyhttp.MiddlewareHandler = (*Markdown)(nil)
	// _ caddyfile.Unmarshaler       = (*Markdown)(nil)
)
