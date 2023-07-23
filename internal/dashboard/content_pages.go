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
	<link href="/builtin/reset.css" rel="stylesheet" />
	<link href="/builtin/tachyon.css" rel="stylesheet" />
	<link href="/builtin/style.css" rel="stylesheet" />
</head>
{{end}}

{{define "index"}}
{{ template "intro" .}}
{{ template "head" . }}
{{ template "body-intro" }}
<div class="flex h-100">
	<section class="w-20">
		<h1>Requests</h2>
		<ul hx-get="/requests" hx-trigger="every 2s" hx-swap="morphdom" style="overflow-y: auto" class="h-100">
		</ul>
	</section>
	<section class="w-80" id="request-inspector">
	</section>
</div>
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
<li id="rid-{{.ID}}"><a href="/inspect-request?rid={{.ID}}" hx-get="/inspect-request?rid={{.ID}}" hx-target="#request-inspector" hx-swap="innerHTML">{{ .Code }} - {{ .URL }}</span></li>
{{- end }}
</ul>
{{end}}

{{define "inspect-request"}}
<dl>
	<dt>URL</dt>
	<dd>{{.URL}}</dd>
	<dt>Request Headers</dt>
	<dd>
		<ul>
			{{range $k, $v := .Request.Headers }}
			<li><strong>{{$k}}</strong>: <span>{{$v}}</span></li>
			{{end}}
		</ul>
	</dd>
	<dt>Request body</dt>
	<dd><pre>{{.Request.Body}}</pre></dd>
	<dt>Response Headers</dt>
	<dd>
		<ul>
			{{range $k, $v := .Request.Headers }}
			<li><strong>{{$k}}</strong>: <span>{{$v}}</span></li>
			{{end}}
		</ul>
	</dd>
	<dt>Response body</dt>
	<dd><pre>{{.Response.Body}}</pre></dd>
</dl>
{{end}}
`
)
