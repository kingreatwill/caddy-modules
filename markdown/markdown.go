package markdown

import (
	"bytes"
	"context"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"path/filepath"
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
	Root       string   `json:"root,omitempty"`
	Template   string   `json:"template,omitempty"`
	Hide       []string `json:"hide,omitempty"`
	MIMETypes  []string `json:"mime_types,omitempty"`
	engine     *convert.MarkdownConvert
	fileSystem fs.FS
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
				engine:     convert.New(),
				fileSystem: osFS{},
			}
		},
	}
}

// Provision sets up the module. #caddy.Provisioner
func (md *Markdown) Provision(ctx caddy.Context) error {
	if md.Root == "" {
		md.Root = "{http.vars.root}"
	}
	// for hide paths that are static (i.e. no placeholders), we can transform them into
	// absolute paths before the server starts for very slight performance improvement
	for i, h := range md.Hide {
		if !strings.Contains(h, "{") && strings.Contains(h, separator) {
			if abs, err := filepath.Abs(h); err == nil {
				md.Hide[i] = abs
			}
		}
	}
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
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	root := repl.ReplaceAll(md.Root, ".")
	filename := strings.TrimSuffix(caddyhttp.SanitizedPathJoin(root, r.URL.Path), "/")
	caddy.Log().Info("ServeHTTP3:", zap.String("path", r.URL.Path), zap.String("root", root), zap.String("filename", filename))
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

// osFS is a simple fs.FS implementation that uses the local
// file system. (We do not use os.DirFS because we do our own
// rooting or path prefixing without being constrained to a single
// root folder. The standard os.DirFS implementation is problematic
// since roots can be dynamic in our application.)
//
// osFS also implements fs.StatFS, fs.GlobFS, fs.ReadDirFS, and fs.ReadFileFS.
type osFS struct{}

func (osFS) Open(name string) (fs.File, error)          { return os.Open(name) }
func (osFS) Stat(name string) (fs.FileInfo, error)      { return os.Stat(name) }
func (osFS) Glob(pattern string) ([]string, error)      { return filepath.Glob(pattern) }
func (osFS) ReadDir(name string) ([]fs.DirEntry, error) { return os.ReadDir(name) }
func (osFS) ReadFile(name string) ([]byte, error)       { return os.ReadFile(name) }

var defaultIndexNames = []string{"index.html", "index.txt"}

const (
	minBackoff, maxBackoff = 2, 5
	separator              = string(filepath.Separator)
)

// Interface guards
var (
	_ caddy.Provisioner           = (*Markdown)(nil)
	_ caddy.Validator             = (*Markdown)(nil)
	_ caddyhttp.MiddlewareHandler = (*Markdown)(nil)
	// _ caddyfile.Unmarshaler       = (*Markdown)(nil)

	_ fs.StatFS     = (*osFS)(nil)
	_ fs.GlobFS     = (*osFS)(nil)
	_ fs.ReadDirFS  = (*osFS)(nil)
	_ fs.ReadFileFS = (*osFS)(nil)
)
