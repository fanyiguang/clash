package route

import (
	"net/http"

	"github.com/Dreamacro/clash/common/crypto"
	"github.com/Dreamacro/clash/config"
	P "github.com/Dreamacro/clash/listener"
	"github.com/Dreamacro/clash/log"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func inbounds() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getInbounds)
	r.Post("/", addInbounds)
	r.Delete("/", deleteInbounds)
	return r
}

func getInbounds(w http.ResponseWriter, r *http.Request) {
	inbounds := P.GetInbounds()
	var result []map[string]interface{}
	for _, inbound := range inbounds {

		sd, err := crypto.AecEcb128Pkcs7EncryptWithDefaultKey([]byte(inbound.GetRawConfigString()))
		if err != nil {
			log.Infoln("Encrypt data err: %s", err)
			sd = ""
		}

		result = append(result, map[string]interface{}{
			"name":       inbound.Name(),
			"type":       inbound.Type(),
			"rawAddress": inbound.RawAddress(),
			"secretData": sd,
		})
	}
	render.JSON(w, r, result)
}

func addInbounds(w http.ResponseWriter, r *http.Request) {
	var params []config.InboundConfig
	err := render.DecodeJSON(r.Body, &params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	err = P.AddInbounds(params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	render.NoContent(w, r)
}

func deleteInbounds(w http.ResponseWriter, r *http.Request) {
	var params []string
	err := render.DecodeJSON(r.Body, &params)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, newError(err.Error()))
		return
	}
	P.DeleteInbounds(params)
	render.NoContent(w, r)
}
