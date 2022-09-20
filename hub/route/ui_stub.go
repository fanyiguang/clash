//go:build without_webui

package route

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func UI(r chi.Router) {
	r.Get("/ui", func(writer http.ResponseWriter, request *http.Request) {
		writer.Write([]byte("web-ui is not included in this build"))
	})
}
