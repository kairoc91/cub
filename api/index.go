package handler

import (
	"net/http"
	"path"
	"strings"
	"sync"

	appAssets "cub"
	"cub/internal/calculator"
	cubRepo "cub/internal/cub"
	"cub/internal/handlers"
	"cub/internal/storage"
)

var (
	appMux   http.Handler
	initOnce sync.Once
	initErr  error
)

func initialize() {
	repo, err := cubRepo.NewRepositoryFromReader(strings.NewReader(appAssets.CUBHistoricoCSV))
	if err != nil {
		initErr = err
		return
	}

	calc := calculator.NewCalculator(repo)
	store, err := storage.NewStorage("") // Modo resiliente/em memória para serverless
	if err != nil {
		initErr = err
		return
	}

	srvHandler, err := handlers.NewServer(repo, calc, store, "static")
	if err != nil {
		initErr = err
		return
	}

	mux := http.NewServeMux()
	srvHandler.RegisterRoutes(mux)
	appMux = mux
}

// Handler é o ponto de entrada executado pelo runtime serverless do Vercel
func Handler(w http.ResponseWriter, r *http.Request) {
	initOnce.Do(initialize)

	if initErr != nil {
		http.Error(w, "Erro ao inicializar simulador CUB: "+initErr.Error(), http.StatusInternalServerError)
		return
	}

	// 1. Tenta recuperar o path original passado pelo rewrite do vercel.json
	reqPath := r.URL.Query().Get("__path")
	if reqPath == "" {
		reqPath = r.Header.Get("x-matched-path")
	}
	if reqPath == "" {
		reqPath = r.URL.Path
	}

	// Limpa e normaliza o path (evita //, ., /api/index, caminhos relativos)
	cleaned := path.Clean("/" + strings.TrimPrefix(reqPath, "/"))
	if cleaned == "/api/index" || cleaned == "/api" || cleaned == "." || cleaned == "" {
		cleaned = "/"
	}
	r.URL.Path = cleaned

	// Remove o parâmetro interno __path para não poluir o restante da aplicação
	if r.URL.RawQuery != "" {
		q := r.URL.Query()
		if q.Has("__path") {
			q.Del("__path")
			r.URL.RawQuery = q.Encode()
		}
	}

	appMux.ServeHTTP(w, r)
}
