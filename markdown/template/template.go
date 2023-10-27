package template

var templates = map[string]string{
	"simple": `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1,minimum-scale=1,maximum-scale=1,user-scalable=no">
<title>{{.Title}}</title>
</head>
<body>
{{.Body}}
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
{{.Body}}
</body>
</html>`,
}