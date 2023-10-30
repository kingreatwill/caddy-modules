package template

import (
	"fmt"
	"io"
	"strings"
	"sync"
	textTemplate "text/template"

	"github.com/kingreatwill/caddy-modules/markdown/convert"
)

var (
	extensions_file_icon   sync.Map // map[string][]string; slice values are append-only.
	extensions_folder_icon sync.Map
)

func init() {
	file_icon_update := map[string][]string{
		"file_type_markdown.svg":   {".md"},
		"file_type_bundler.svg":    {"gemfile"},
		"file_type_zip.svg":        {".gz", ".7z", ".tar", ".tgz", ".bz"},
		"file_type_go_package.svg": {".mod", ".sum"},
		"file_type_docker2.svg":    {"dockerfile"},
		"file_type_image.svg":      {".jpeg", ".jpg", ".gif", ".png", ".bmp", ".tiff", ".ico"},
	}
	folder_icon_update := map[string][]string{
		"folder_type_windows.svg": {"win"},
		"folder_type_tests.svg":   {"test", "integration", "specs", "spec"},
		"folder_type_images.svg":  {"img", "image", "imgs"},
		"folder_type_src.svg":     {"source", "sources"},
		"folder_type_log.svg":     {"logs"},
		"folder_type_locale.svg":  {"lang", "language", "languages", "locales", "internationalization", "i18n", "globalization", "g11n", "localization", "l10n"},
	}
	for icon, v := range file_icon_update {
		for _, key := range v {
			extensions_file_icon.Store(key, icon)
		}
	}
	for icon, v := range folder_icon_update {
		for _, key := range v {
			extensions_folder_icon.Store(key, icon)
		}
	}
}

func GetExtensionsIcon(ext string, isdir bool) string {
	ext = strings.ToLower(ext)
	if isdir {
		if v, ok := extensions_folder_icon.Load(ext); ok {
			return fmt.Sprint(v)
		}
		return "default_folder.svg"
	}
	if v, ok := extensions_file_icon.Load(ext); ok {
		return fmt.Sprint(v)
	}
	return "default_file.svg"
}

var Templates = map[string]string{
	"simple": `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1,maximum-scale=1,user-scalable=no">
<title>{{.Title}}</title>
</head>
<body>
{{.MdHtml}}
</body>
</html>
`,
	"normal": `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1,maximum-scale=1,user-scalable=no">
<title>{{.Title}}</title>
<style>
body {
	font-size:11pt;
}

h1 { font-size:1.8em; }
h2 { font-size:1.5em; }
h3 { font-size:1.3em; }
h4 { font-size:1.1em; }

p, ul, ol {
	margin-block-start:0.5em;
	margin-block-end:0.5em;
	line-height:1.5em;
}
ul, ol {
	padding-left:1.5em;
	padding-inline-start:1.5em;
}

table {
	border-collapse:collapse;
}
th, td {
	border:1px solid gray;
	padding:0.2em 0.5em;
}
thead tr:nth-child(odd),
tbody tr:nth-child(even) {
	background-color:#f0f0f0;
}

code {
	display:inline-block;
	background-color:#e8e8e8;
	padding:0.1em 0.2em;
	margin:0 0.1em;
	border-radius:0.2em;
	text-decoration:inherit;
	line-height:1.2em;
}
pre code {
	max-width:98%;
	overflow-x:auto;
	background-color:#f6f8fa;
	padding:0.8em 0.5em;
	font-size:10pt;
}
warn {
	color:red;
}
footnote {
	display:block;
	font-size:0.6em;
	margin-top:4em;
}
footnote * {
	font-size:0.6em;
}
</style>
</head>
<body>
{{.MdHtml}}
</body>
</html>`,
}


func Execute(wr io.Writer, tmplStr string,data *convert.TemplateData) error {
	// 解析模板
	tmpl, err := textTemplate.New("markdown").Parse(tmplStr); 
	if err != nil {
		return err
	}
	return tmpl.Execute(wr, data)
}