// Package web serves the platform's single-page application.
//
// The build output is embedded into the binary, so a deployment is one
// executable with one data directory and no static-file server to configure.
package web

import (
	"embed"
	"io/fs"
	"net/http"
	"path"
	"strings"
)

//go:embed all:dist
var distFS embed.FS

// Assets returns the built frontend assets.
//
// The directory is empty in a source checkout that has not been built yet.
// The server detects that and serves the placeholder page instead of
// failing, so the API stays usable during backend development.
func Assets() (fs.FS, bool) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, false
	}
	entries, err := fs.ReadDir(sub, ".")
	if err != nil {
		return nil, false
	}
	for _, e := range entries {
		if e.Name() == "index.html" {
			return sub, true
		}
	}
	return sub, false
}

// Handler returns the handler that serves the single-page application.
//
// Unknown paths fall back to index.html, because the application owns the
// routing once it has loaded.
func Handler() http.Handler {
	assets, built := Assets()
	if !built {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(placeholderHTML))
		})
	}
	fileServer := http.FileServer(http.FS(assets))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" {
			name = "index.html"
		}
		if _, err := fs.Stat(assets, name); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

// placeholderHTML is served when the frontend has not been built.
//
// It tells the operator what to run instead of showing a blank page, because
// a blank page at the root of an evaluation platform looks like a broken
// deployment.
const placeholderHTML = `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>b-eval</title></head>
<body style="font-family: system-ui, sans-serif; margin: 3rem auto; max-width: 42rem; line-height: 1.6;">
<h1>b-eval</h1>
<p>The API is running, but the frontend has not been built.</p>
<pre style="background:#f4f4f5;padding:1rem;border-radius:6px;">cd web &amp;&amp; npm install &amp;&amp; npm run build</pre>
<p>The API is available under <code>/api</code>. Try <a href="/api/healthz">/api/healthz</a>.</p>
</body>
</html>
`
