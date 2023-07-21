package dashboard

const (
	pages = `
{{define "intro"}}
<!doctype html>
<html>
{{end}}

{{define "otro"}}
</html>
{{end}}

{{define "head" }}
<head>
	<title>{{.Title}}</title>
	<script src="/builtin/htmx.js"></script>
	<link href="/builtin/style.css" rel="stylesheet" content-type="text/css">
</head>
{{end}}

{{define "index"}}
{{ template "intro" .}}
{{ template "head" . }}
<body>
<section>
	<h1>Requests</h2>
	<ul hx-get="/requests" hx-trigger="every 2s">
	</ul>
</section>
</body>
{{ template "otro" .}}
{{end}}

{{define "requests" }}
{{ range .Requests -}}
<li id="rid-{{.ID}}">{{ .Code }} - {{ .URL }}</li>
{{- end }}
{{end}}
`
)
