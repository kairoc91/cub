package handler

import (
	"net/http"
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
	if p := r.URL.Query().Get("__path"); p != "" {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		r.URL.Path = p
	} else if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" {
		// 2. Fallback para header de rewrite
		if matchedPath != "/api/index" && matchedPath != "/api" {
			r.URL.Path = matchedPath
		}
	}

	// Normaliza rotas que apontem diretamente para o bundle da função
	if r.URL.Path == "/api/index" || r.URL.Path == "/api" || r.URL.Path == "" {
		r.URL.Path = "/"
	}

	appMux.ServeHTTP(w, r)
}
