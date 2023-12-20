package search

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/blevesearch/bleve/v2"
	"github.com/blevesearch/bleve/v2/search/query"
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
	index    bleve.Index
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
	index, err := CreateIndex()
	if err != nil {
		sch.logger.Debug("watch CreateIndex Error", zap.Error(err))
	}
	sch.index = index
	// 监听文件变化
	repl := ctx.Value(caddy.ReplacerCtxKey).(*caddy.Replacer)
	root := repl.ReplaceAll(sch.Root, ".")
	sch.watch = NewNotifyFile(sch.logger, sch.IndexDoc)
	go sch.watch.WatchDir(root)
	return nil
}

// Validate ensures md has a valid configuration. #caddy.Validator
// Validate should be a read-only function. It is run after the Provision() method.
func (sch *Search) Validate() error {
	return nil
}

// ServeHTTP implements #caddyhttp.MiddlewareHandler.
func (sch *Search) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {
	return nil
}

// IndexDoc 索引文件
func (sch *Search) IndexDoc(path string, remove bool) {
	found, err := regexp.MatchString(sch.Regexp, path)
	if err != nil {
		sch.logger.Debug("regexp Error",
			zap.String("path", path),
			zap.Error(err))
		return
	}
	if !found {
		return
	}
	key := path
	if remove {
		if err := sch.index.Delete(key); err != nil {
			sch.logger.Debug("watch Delete Doc Error",
				zap.String("key", key),
				zap.Error(err))
		}
		return
	}
	bytes, err := os.ReadFile(path)
	if err != nil {
		sch.logger.Debug("watch index doc ReadFile Error",
			zap.String("key", key),
			zap.String("path", path),
			zap.Error(err))
		return
	}
	message := struct {
		Id   string
		Path string
		Body string
	}{
		Id:   key,
		Path: path,
		Body: string(bytes),
	}
	err = sch.index.Index(message.Id, message)
	if err != nil {
		sch.logger.Debug("watch Index Doc Error",
			zap.String("key", key),
			zap.Error(err))
		return
	}
}

// Search 搜索文件
func (sch *Search) Search(queryStr string) map[string][]string {
	var q query.Query
	q = bleve.NewQueryStringQuery(queryStr)
	phrases := strings.Split(queryStr, " ")
	if len(phrases) > 1 {
		q = bleve.NewPhraseQuery(phrases, "Body")
	}
	searchRequest := bleve.NewSearchRequest(q)
	searchRequest.Highlight = bleve.NewHighlight()
	searchResult, err := sch.index.Search(searchRequest)
	if err != nil {
		sch.logger.Debug("Search Error",
			zap.String("query", queryStr),
			zap.Error(err))
		return nil
	}
	if searchResult.Total == 0 {
		return nil
	}
	result := make(map[string][]string, len(searchResult.Hits))
	for _, re := range searchResult.Hits {
		result[re.ID] = re.Fragments["Body"]
	}
	return result
}

// Interface guards
var (
	_ caddy.Provisioner           = (*Search)(nil)
	_ caddy.Validator             = (*Search)(nil)
	_ caddyhttp.MiddlewareHandler = (*Search)(nil)
)
