package api

import (
	"net/http"
)

func RegisterAllApi() {
	http.HandleFunc("/api/version", apiGetVersion)
	http.HandleFunc("/api/conf", apiGetConfig)
	http.HandleFunc("/api/restart", apiRestart)
}
