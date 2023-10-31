package markdown

import (
	"bytes"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
	"github.com/kingreatwill/caddy-modules/markdown/convert"
	"github.com/kingreatwill/caddy-modules/markdown/template"
)

type Markdown struct {
	Root      string   `json:"root,omitempty"`
	Template  string   `json:"template,omitempty"`
	Hide      []string `json:"hide,omitempty"`
	MIMETypes []string `json:"mime_types,omitempty"`
	// The names of files to try as index files if a folder is requested.
	// Default: index.html index.htm
	IndexNames []string `json:"index,omitempty"`
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
	if md.MIMETypes == nil {
		md.MIMETypes = []string{"text/markdown"}
	}
	if md.IndexNames == nil {
		md.IndexNames = defaultIndexNames
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
func (md *Markdown) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {

	// repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	// root := repl.ReplaceAll(md.Root, ".")
	// filename := strings.TrimSuffix(caddyhttp.SanitizedPathJoin(root, r.URL.Path), "/")

	// if !strings.HasSuffix(_path, ".md") && !strings.HasSuffix(_path, ".markdown") {
	// 	return next.ServeHTTP(w, r)
	// }

	// info, err := fs.Stat(md.fileSystem, filename)
	// if err != nil {
	// 	return err
	// }

	// caddy.Log().Info("ServeHTTP3:",
	// 	zap.String("path", r.URL.Path),
	// 	zap.String("root", root),
	// 	zap.String("filename", filename),
	// 	zap.Bool("IsDir", info.IsDir()))

	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	shouldBuf := func(status int, header http.Header) bool {
		if strings.HasSuffix(r.URL.Path, ".md") || strings.HasSuffix(r.URL.Path, ".markdown") {
			return true
		}
		ct := header.Get("Content-Type")
		for _, mt := range md.MIMETypes {
			if strings.Contains(ct, mt) {
				return true
			}
		}
		return false
	}
	rec := caddyhttp.NewResponseRecorder(w, buf, shouldBuf)
	err = next.ServeHTTP(rec, r)
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
		root := md.getRoot(r)
		fs := http.Dir(root)
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
	html, err := md.renderMarkdown(r, inputStr, tmpl)
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

func (md *Markdown) renderMarkdown(r *http.Request, inputStr, tmplStr string) (string, error) {
	// 获取目录数据
	data, err := md.getTemplateData(r)
	if err != nil {
		return "", err
	}
	// 转换
	err = md.engine.Convert(inputStr, data)
	if err != nil {
		return "", err
	}
	if data.Title == "" {
		orignalRequest := r.Context().Value(caddyhttp.OriginalRequestCtxKey).(http.Request)
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

func (md *Markdown) getRoot(r *http.Request) string {
	repl := r.Context().Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	return repl.ReplaceAll(md.Root, ".")
}

func (md *Markdown) getTemplateData(r *http.Request) (data *convert.TemplateData, err error) {
	data = &convert.TemplateData{
		CurrentDirs: []convert.TemplateFileItemData{},
	}
	root := md.getRoot(r)
	filename := strings.TrimSuffix(caddyhttp.SanitizedPathJoin(root, r.URL.Path), "/")

	info, err := fs.Stat(md.fileSystem, filename)
	if err != nil {
		return nil, err
	}

	if filename != "." {
		data.UpperPath = strings.ReplaceAll(filepath.Dir(filename), "\\", "/")
		if !strings.HasSuffix(data.UpperPath, "/") {
			data.UpperPath = data.UpperPath + "/"
		}
		if !strings.HasPrefix(data.UpperPath, "/") {
			data.UpperPath = "/" + data.UpperPath
		}
	}

	listDir := filename
	data.CurrentFile = filename
	data.CurrentIsFile = !info.IsDir()

	if !info.IsDir() {
		listDir = filepath.Dir(filename)
	}
	for _, fi := range md.listdir(listDir) {
		if strings.HasPrefix(fi.Name(), ".") || strings.HasPrefix(fi.Name(), "_") {
			continue
		}
		item := convert.TemplateFileItemData{
			Name:          fi.Name(),
			IsFile:        !fi.IsDir(),
			FileExtension: fi.Name(),
			Href:          strings.ReplaceAll(path.Join(listDir, fi.Name()), "\\", "/"),
		}
		if !fi.IsDir() {
			item.FileExtension = path.Ext(fi.Name())
			if info.IsDir() {
				for _, index := range md.IndexNames {
					if index == fi.Name() {
						data.CurrentFile = item.Href
						data.CurrentIsFile = true
					}
				}
			}
		} else {
			item.Href = item.Href + "/"
		}
		if !strings.HasPrefix(item.Href, root) {
			item.Href = strings.Replace(item.Href, root, "", 1)
		}
		if !strings.HasPrefix(item.Href, "/") {
			item.Href = "/" + item.Href
		}
		//item.Icon = template.GetExtensionsIcon(item.FileExtension, fi.IsDir())
		data.CurrentDirs = append(data.CurrentDirs, item)
	}
	sort.Slice(data.CurrentDirs, func(i, j int) bool {
		// 1. IsFile:升序排序
		if data.CurrentDirs[i].IsFile != data.CurrentDirs[j].IsFile {
			return data.CurrentDirs[j].IsFile
		}
		// 2. Name:升序排序
		return strings.ToLower(data.CurrentDirs[i].Name) < strings.ToLower(data.CurrentDirs[j].Name)
	})
	return
}

func (md *Markdown) listdir(pathname string) (fileInfos []fs.FileInfo) {
	dirEntries, err := os.ReadDir(pathname)
	if err != nil {
		log.Println(err)
		return nil
	}
	for _, de := range dirEntries {
		fi, _ := de.Info()
		fileInfos = append(fileInfos, fi)
	}
	return
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

var defaultIndexNames = []string{"README.md", "README.markdown", "readme.markdown", "readme.md"}

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
