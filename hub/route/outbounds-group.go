package route

import (
	"net/http"

	"github.com/Dreamacro/clash/tunnel"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func outboundGroups() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getOutbounds)
	r.Post("/", addOutboudGroups)
	r.Delete("/", deleteOutbounds)
	return r
}

func addOutboudGroups(w http.ResponseWriter, r *http.Request) {
	var params []map[string]any
	err := render.DecodeJSON(r.Body, &params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	err = tunnel.AddOutboundGroups(params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	render.NoContent(w, r)
}
