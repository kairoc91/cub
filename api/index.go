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

	// Restaura o path original da URL quando reescrito pelo Vercel
	if matchedPath := r.Header.Get("x-matched-path"); matchedPath != "" {
		r.URL.Path = matchedPath
	}

	appMux.ServeHTTP(w, r)
}
