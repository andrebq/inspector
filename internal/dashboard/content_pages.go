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
	<script src="/builtin/morphdom.js"></script>
	<link href="/builtin/style.css" rel="stylesheet" content-type="text/css">
</head>
{{end}}

{{define "index"}}
{{ template "intro" .}}
{{ template "head" . }}
{{ template "body-intro" }}
<section>
	<h1>Requests</h2>
	<ul hx-get="/requests" hx-trigger="every 2s" hx-swap="morphdom">
	</ul>
</section>
{{ template "body-otro" }}
{{ template "otro" .}}
{{end}}

{{ define "body-intro" }}
<body hx-ext="morphdom-swap">
{{ end }}

{{ define "body-otro"}}
</body>
{{ end }}

{{define "requests" }}
<ul hx-get="/requests" hx-trigger="every 2s" hx-swap="morphdom">
{{ range .Requests -}}
<li id="rid-{{.ID}}"><a href="#inspect">{{ .Code }} - {{ .URL }}</a></li>
{{- end }}
</ul>
{{end}}

{{define "inspect-request"}}
<dl>
	<dt>URL</dt>
	<dd>{{.URL}}</dd>
	<dt>Host</dt>
	<dd>{{.Host}}</dd>
	<dt>Request Headers</dt>
	<dd>
		<ul>
			{{range $k, $v := .Request.Headers }}
			<strong>{{$k}}</strong>: <span>{{$v}}</span>
			{{end}}
		</ul>
	</dd>
</dl>
{{end}}
`
)
