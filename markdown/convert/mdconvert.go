package convert

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	katex "github.com/kingreatwill/goldmark-katex"
	"github.com/yuin/goldmark"
	emoji "github.com/yuin/goldmark-emoji"
	highlighting "github.com/yuin/goldmark-highlighting"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
	"go.abhg.dev/goldmark/toc"
)

type MarkdownConvert struct {
	engine goldmark.Markdown
}

func New() *MarkdownConvert {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
			katex.KaTeX,
			emoji.Emoji,
			//mathjax.MathJax,
			highlighting.Highlighting,
			&toc.Extender{},
			&mermaid.Extender{},
			meta.Meta,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)
	return &MarkdownConvert{
		engine: md,
	}
}

type TemplateData struct {
	SiteUrl     string `remark:"站点地址"`
	Title       string `remark:"<title>Title"`
	Keywords    string `remark:"<meta>Keywords逗号隔开"`
	Description string `remark:"<meta>description"`
	HasKatex    bool   `remark:"md中是否解析了katex"`
	HasMermaid  bool   `remark:"md中是否解析了mermaid"`

	CurrentDirs   []TemplateFileItemData `remark:"路径"`
	CurrentFile   string                 `remark:"当前渲染文件(也有可能是目录)"`
	CurrentIsFile bool                   `remark:"是否有渲染文件"`
	Content       []byte                 `remark:"md"`
	MdHtml        string                 `remark:"html"`
	UpperPath     string                 `remark:"上一级连接"`

	GitStartDate string `remark:"一年前的日期"`
	GitStatsData string `remark:"截止到目前为止所有的每日提交数量"`
}

type TemplateFileItemData struct {
	FileExtension string `remark:"文件后缀名"`
	IsFile        bool   `remark:"是否文件"`
	Name          string `remark:"文件或目录名"`
	Href          string `remark:"连接,不带SiteUrl"`
	Icon          string `remark:"Icon"`
}

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func (c *MarkdownConvert) Convert(mdStr string, data *TemplateData) (err error) {

	data.Content = []byte(mdStr)

	// var buf bytes.Buffer
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	context := parser.NewContext()
	if err = c.engine.Convert(data.Content, buf, parser.WithContext(context)); err != nil {
		return err
	}

	data.MdHtml = buf.String()

	metaData := meta.Get(context)
	if value, ok := metaData["Title"]; ok {
		data.Title = fmt.Sprintf("%v", value)
	} else if value, ok := metaData["title"]; ok {
		data.Title = fmt.Sprintf("%v", value)
	}

	keywordsFunc := func(value interface{}) {
		if newValue, ok := value.([]interface{}); ok {
			var tags []string
			for _, tag := range newValue {
				tags = append(tags, fmt.Sprintf("%v", tag))
			}
			data.Keywords = strings.Join(tags, ",")
		} else {
			data.Keywords = fmt.Sprintf("%v", value)
		}
	}
	if value, ok := metaData["Keywords"]; ok {
		keywordsFunc(value)
	} else if value, ok := metaData["keywords"]; ok {
		keywordsFunc(value)
	} else if value, ok := metaData["Tags"]; ok {
		keywordsFunc(value)
	} else if value, ok := metaData["tags"]; ok {
		keywordsFunc(value)
	}

	if value, ok := metaData["Description"]; ok {
		data.Description = fmt.Sprintf("%v", value)
	} else if value, ok := metaData["description"]; ok {
		data.Description = fmt.Sprintf("%v", value)
	} else if value, ok := metaData["Summary"]; ok {
		data.Description = fmt.Sprintf("%v", value)
	} else if value, ok := metaData["summary"]; ok {
		data.Description = fmt.Sprintf("%v", value)
	}
	return nil
}
