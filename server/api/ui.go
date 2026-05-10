package api

import (
	"io/fs"
	"net/http"
	"strings"

	angelixweb "github.com/YumikoKawaii/angelix/web"
)

func newUIHandler() http.Handler {
	dist, err := fs.Sub(angelixweb.FS, "dist")
	if err != nil {
		panic("embedded web/dist not found: " + err.Error())
	}
	files := http.FileServer(http.FS(dist))
	return &uiHandler{dist: dist, files: files}
}

type uiHandler struct {
	dist  fs.FS
	files http.Handler
}

func (h *uiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// If the file exists in the embedded FS, serve it directly.
	// Otherwise serve index.html so React Router handles the path.
	if _, err := h.dist.Open(path); err != nil {
		r = r.Clone(r.Context())
		r.URL.Path = "/"
	}
	h.files.ServeHTTP(w, r)
}
