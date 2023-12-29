package search

import (
	"encoding/json"
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
		sch.Root = "."
	}
	if sch.Endpoint == "" {
		sch.Endpoint = "/search"
	}
	if sch.Regexp == "" {
		sch.Regexp = ".*.md"
	}
	index, err := CreateIndex()
	if err != nil {
		sch.logger.Debug("watch CreateIndex Error", zap.Error(err))
	}
	sch.index = index
	// 监听文件变化
	sch.logger.Debug("Provision", zap.String("root", sch.Root))
	sch.watch = NewNotifyFile(sch.logger, sch.IndexDoc)
	go sch.watch.WatchDir(sch.Root)
	return nil
}

// Validate ensures md has a valid configuration. #caddy.Validator
// Validate should be a read-only function. It is run after the Provision() method.
func (sch *Search) Validate() error {
	return nil
}

// ServeHTTP implements #caddyhttp.MiddlewareHandler.
func (sch *Search) ServeHTTP(w http.ResponseWriter, r *http.Request, next caddyhttp.Handler) (err error) {
	if r.URL.Path == sch.Endpoint {
		if r.Header.Get("Accept") == "application/json" {
			return sch.SearchJSON(w, r)
		}
		return sch.SearchHTML(w, r)
	}
	return next.ServeHTTP(w, r)
}

type Result struct {
	Path string
	Body []string
}

func (sch *Search) searchResult(r *http.Request) []Result {
	q := r.URL.Query().Get("q")
	indexResult := sch.Search(q)
	results := make([]Result, 0, len(indexResult))
	for key, value := range indexResult {
		results = append(results, Result{
			Path: key,
			Body: value,
		})
	}
	return results
}

// SearchJSON renders the search results in JSON format
func (sch *Search) SearchJSON(w http.ResponseWriter, r *http.Request) error {
	results := sch.searchResult(r)
	jresp, err := json.Marshal(results)
	if err != nil {
		return err
	}
	_, err = w.Write(jresp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return nil
}

// SearchHTML renders the search results in the HTML template
func (sch *Search) SearchHTML(w http.ResponseWriter, r *http.Request) error {
	results := sch.searchResult(r)
	jresp, err := json.Marshal(results)
	if err != nil {
		return err
	}
	_, err = w.Write(jresp)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	return nil

	//indexResult := s.Indexer.Search(q)
	//
	//results := make([]Result, len(indexResult))
	//
	//for i, result := range indexResult {
	//	results[i] = Result{
	//		Path:     result.Path(),
	//		Title:    result.Title(),
	//		Modified: result.Modified(),
	//		Body:     string(result.Body()),
	//	}
	//}
	//
	//qresults := QueryResults{
	//	Context: httpserver.Context{
	//		Root: http.Dir(s.SiteRoot),
	//		Req:  r,
	//		URL:  r.URL,
	//	},
	//	Query:   q,
	//	Results: results,
	//}
	//
	//var buf bytes.Buffer
	//err := s.Config.Template.Execute(&buf, qresults)
	//if err != nil {
	//	return http.StatusInternalServerError, err
	//}
	//w.Header().Set("Content-Type", "text/html; charset=utf-8")
	//
	//buf.WriteTo(w)
	//return http.StatusOK, nil
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
