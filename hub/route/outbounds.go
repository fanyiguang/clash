package route

import (
	"net/http"

	"github.com/Dreamacro/clash/tunnel"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func outbounds() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getOutbounds)
	r.Post("/", addOutbounds)
	r.Delete("/", deleteOutbounds)
	return r
}

func getOutbounds(w http.ResponseWriter, r *http.Request) {
	proxies := tunnel.Proxies()
	render.JSON(w, r, render.M{
		"proxies": proxies,
	})
}

func addOutbounds(w http.ResponseWriter, r *http.Request) {
	var params []map[string]any
	err := render.DecodeJSON(r.Body, &params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	err = tunnel.AddOutbounds(params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	render.NoContent(w, r)
}

func deleteOutbounds(w http.ResponseWriter, r *http.Request) {
	var params []string
	err := render.DecodeJSON(r.Body, &params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	tunnel.DeleteOutbounds(params)
	render.NoContent(w, r)
}
