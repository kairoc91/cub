package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cub/internal/calculator"
	"cub/internal/cub"
	"cub/internal/handlers"
	"cub/internal/storage"
)

func main() {
	port := flag.String("port", "8080", "Porta HTTP do servidor")
	csvPath := flag.String("csv", "cub_historico.csv", "Caminho do arquivo CSV com histórico do CUB")
	storagePath := flag.String("storage", "data/propostas.json", "Caminho do arquivo de persistência de propostas")
	staticDir := flag.String("static", "static", "Diretório de arquivos estáticos")
	flag.Parse()

	log.Println("🚀 Iniciando Simulador CUB (Go + HTMX + Tailwind v4)...")

	// 1. Repositório CUB
	inicioCUB := time.Now()
	repo, err := cub.NewRepositoryFromFile(*csvPath)
	if err != nil {
		log.Fatalf("❌ Falha crítica ao carregar CUB histórico: %v", err)
	}
	log.Printf("✔ CUB Histórico indexado em %v (%d registros). CUB mais recente: %s (%s)",
		time.Since(inicioCUB),
		len(repo.GetHistorico()),
		repo.GetMaisRecente().MesAno,
		repo.GetMaisRecente().ValorFormatado,
	)

	// 2. Motor Financeiro
	calc := calculator.NewCalculator(repo)

	// 3. Persistência
	store, err := storage.NewStorage(*storagePath)
	if err != nil {
		log.Fatalf("❌ Falha ao inicializar armazenamento: %v", err)
	}
	log.Printf("✔ Armazenamento inicializado. %d propostas gravadas.", len(store.List()))

	// 4. Servidor HTTP & Handlers
	srvHandler, err := handlers.NewServer(repo, calc, store, *staticDir)
	if err != nil {
		log.Fatalf("❌ Falha ao compilar templates HTML: %v", err)
	}

	mux := http.NewServeMux()
	srvHandler.RegisterRoutes(mux)

	server := &http.Server{
		Addr:         ":" + *port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Canal para shutdown gracioso
	parada := make(chan os.Signal, 1)
	signal.Notify(parada, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Printf("\n✨ Servidor rodando com sucesso em http://localhost:%s\n\n", *port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Erro fatal no servidor HTTP: %v", err)
		}
	}()

	<-parada
	log.Println("\n🛑 Encerrando servidor de forma segura...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Erro durante shutdown: %v", err)
	}
	log.Println("✔ Servidor finalizado.")
}
