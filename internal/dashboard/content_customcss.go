package dashboard

const (
	customCSS = `
html, body {
	padding: 0;
	margin: 0;
	height: 100vh;
	overflow: hidden;
}

body {
	font-family: sans-serif;
	background-color: #fffceb;
}

li {
	padding: 0.5rem;
}

dd {
	margin-left: 2em;
}

dt {
	font-weight: bold;
	font-size: 110%;
}

dd ul {
	list-style-type: none;
}

pre.limit-h {
	max-height: 50em;
	overflow: auto;
}

.vflow {
	display: flex;
	flex-direction: column;
	height: 100%;
	box-sizing: border-box;
}

hr {
	margin: 1.5rem 0 1.5rem 0;
}

.pill {
	margin: 0.5rem;
}

pre {
	border-left: solid 0.5rem #96ccff;
	padding-left: 1rem;
}
	`
)
