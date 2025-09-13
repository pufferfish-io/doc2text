package api

import (
    "net/http"
)

type Options struct {
    HealthCheckPath string
}

func NewRouter(opt Options) *http.ServeMux {
    mux := http.NewServeMux()

    path := opt.HealthCheckPath
    if path == "" {
        path = "/healthz"
    }
    mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    return mux
}
