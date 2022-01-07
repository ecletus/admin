package admin

import (
	"net/http"

	"github.com/moisespsena-go/httpu"
)

func HttpRedirectFrame(w http.ResponseWriter, r *http.Request, url string, status int) {
	if httpu.IsXhrRequest(r) {
		w.Header().Set("X-Location", url)
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, url, status)
}
