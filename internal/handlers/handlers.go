package handlers

import (
	"html/template"
	"io/fs"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	appAssets "cub"
	"cub/internal/calculator"
	"cub/internal/cub"
	"cub/internal/models"
	"cub/internal/storage"
)

// Server encapsulates dependencies and routes
type Server struct {
	repo       *cub.Repository
	calc       *calculator.Calculator
	store      *storage.Storage
	tmpl       *template.Template
	staticPath string
}

// NewServer creates a new HTTP server handler instance
func NewServer(repo *cub.Repository, calc *calculator.Calculator, store *storage.Storage, staticPath string) (*Server, error) {
	tmpl, err := loadTemplates()
	if err != nil {
		return nil, err
	}

	return &Server{
		repo:       repo,
		calc:       calc,
		store:      store,
		tmpl:       tmpl,
		staticPath: staticPath,
	}, nil
}

func loadTemplates() (*template.Template, error) {
	baseDirs := []string{"templates", "../../templates", "../templates"}
	selectedDir := ""
	for _, dir := range baseDirs {
		if _, err := os.Stat(dir); err == nil {
			selectedDir = dir
			break
		}
	}
	if selectedDir != "" {
		tmpl := template.New("base")
		var err error
		tmpl, err = tmpl.ParseGlob(selectedDir + "/*.html")
		if err == nil {
			partialsDir := selectedDir + "/partials"
			if _, errStat := os.Stat(partialsDir); errStat == nil {
				_, _ = tmpl.ParseGlob(partialsDir + "/*.html")
			}
			return tmpl, nil
		}
	}

	// Fallback para templates embutidos (Vercel Serverless / produção isolada)
	return template.ParseFS(appAssets.TemplatesFS, "templates/*.html", "templates/partials/*.html")
}

// RegisterRoutes binds all application endpoints to the mux
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Arquivos estáticos (CSS, JS) com fallback para filesystem embutido
	var staticFS http.FileSystem
	if _, err := os.Stat(s.staticPath); err == nil {
		staticFS = http.Dir(s.staticPath)
	} else if sub, err := fs.Sub(appAssets.StaticFS, "static"); err == nil {
		staticFS = http.FS(sub)
	} else {
		staticFS = http.FS(appAssets.StaticFS)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(staticFS)))

	// Favicon rápido (evita renderizar index no favicon)
	mux.HandleFunc("GET /favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// Rota Principal (SSR) - catch-all GET / para garantir que / e variações renderizem a home
	mux.HandleFunc("GET /", s.handleIndex)
	mux.HandleFunc("GET /api/index", s.handleIndex)

	// Rotas da API / HTMX
	mux.HandleFunc("POST /api/calcular", s.handleCalcular)
	mux.HandleFunc("POST /api/propostas", s.handleSalvarProposta)
	mux.HandleFunc("GET /api/propostas", s.handleListarPropostas)
	mux.HandleFunc("DELETE /api/propostas", s.handleLimparPropostas)
	mux.HandleFunc("DELETE /api/propostas/{id}", s.handleExcluirProposta)
	mux.HandleFunc("GET /api/propostas/{id}/carregar", s.handleCarregarProposta)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	maisRecente := s.repo.GetMaisRecente()

	// Input padrão inicial configurado com o exemplo didático (R$ 550.000 / CUB Ago-2023)
	input := models.SimulacaoInput{
		Titulo:             "Exemplo Didático • Imóvel R$ 550.000 (CUB 08/2023)",
		CUBBase:            "2023-08",
		ValorSinal:         55000,
		ValorEntrada:       110000,
		TipoParcela:        "mensal",
		QtdParcelas:        24,
		PrimeiroVencimento: "2023-09-10",
		ValorParcela:       13750,
		ValorChaves:        55000,
	}

	resultado := s.calc.Calcular(input)
	propostas := s.store.List()

	data := models.PageData{
		CUBHistorico:   s.repo.GetHistorico(),
		CUBMaisRecente: maisRecente,
		Input:          input,
		Resultado:      resultado,
		Propostas:      propostas,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Erro ao renderizar página: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleCalcular(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form inválido", http.StatusBadRequest)
		return
	}

	input := s.parseFormInput(r)
	resultado := s.calc.Calcular(input)

	data := struct {
		Input     models.SimulacaoInput
		Resultado models.SimulacaoResultado
	}{
		Input:     input,
		Resultado: resultado,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "results.html", data); err != nil {
		http.Error(w, "Erro ao renderizar cálculo: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleSalvarProposta(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form inválido", http.StatusBadRequest)
		return
	}

	input := s.parseFormInput(r)
	if input.Titulo == "" {
		input.Titulo = "Proposta " + time.Now().Format("02/01 - 15:04")
	}

	resultado := s.calc.Calcular(input)
	_, err := s.store.Save(input, resultado)
	if err != nil {
		http.Error(w, "Erro ao salvar proposta: "+err.Error(), http.StatusInternalServerError)
		return
	}

	propostas := s.store.List()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", `{"showToast": "Proposta salva com sucesso no servidor!"}`)
	if err := s.tmpl.ExecuteTemplate(w, "proposals.html", propostas); err != nil {
		http.Error(w, "Erro ao renderizar propostas: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleListarPropostas(w http.ResponseWriter, r *http.Request) {
	propostas := s.store.List()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "proposals.html", propostas); err != nil {
		http.Error(w, "Erro ao renderizar propostas: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleExcluirProposta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id != "" {
		_ = s.store.Delete(id)
	}

	propostas := s.store.List()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", `{"showToast": "Proposta removida com sucesso."}`)
	if err := s.tmpl.ExecuteTemplate(w, "proposals.html", propostas); err != nil {
		http.Error(w, "Erro ao renderizar propostas: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleLimparPropostas(w http.ResponseWriter, r *http.Request) {
	_ = s.store.ClearAll()
	propostas := s.store.List()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("HX-Trigger", `{"showToast": "Todas as propostas foram limpas."}`)
	if err := s.tmpl.ExecuteTemplate(w, "proposals.html", propostas); err != nil {
		http.Error(w, "Erro ao renderizar propostas: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleCarregarProposta(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	proposta, found := s.store.Get(id)
	if !found {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	resultado := s.calc.Calcular(proposta.Input)
	data := models.PageData{
		CUBHistorico:   s.repo.GetHistorico(),
		CUBMaisRecente: s.repo.GetMaisRecente(),
		Input:          proposta.Input,
		Resultado:      resultado,
		Propostas:      s.store.List(),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, "Erro ao renderizar proposta: "+err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) parseFormInput(r *http.Request) models.SimulacaoInput {
	qtd, _ := strconv.Atoi(r.FormValue("qtd_parcelas"))
	if qtd <= 0 {
		qtd = 1
	}

	return models.SimulacaoInput{
		Titulo:             strings.TrimSpace(r.FormValue("titulo")),
		CUBBase:            strings.TrimSpace(r.FormValue("cub_base")),
		ValorSinal:         parseCurrency(r.FormValue("valor_sinal")),
		ValorEntrada:       parseCurrency(r.FormValue("valor_entrada")),
		TipoParcela:        strings.TrimSpace(r.FormValue("tipo_parcela")),
		QtdParcelas:        qtd,
		PrimeiroVencimento: strings.TrimSpace(r.FormValue("primeiro_vencimento")),
		ValorParcela:       parseCurrency(r.FormValue("valor_parcela")),
		ValorChaves:        parseCurrency(r.FormValue("valor_chaves")),
	}
}

func parseCurrency(valStr string) float64 {
	valStr = strings.TrimSpace(valStr)
	if valStr == "" {
		return 0
	}
	valStr = strings.ReplaceAll(valStr, "R$", "")
	valStr = strings.ReplaceAll(valStr, " ", "")
	valStr = strings.ReplaceAll(valStr, "\u00a0", "") // non-breaking space

	if strings.Contains(valStr, ".") && strings.Contains(valStr, ",") {
		valStr = strings.ReplaceAll(valStr, ".", "")
		valStr = strings.ReplaceAll(valStr, ",", ".")
	} else if strings.Contains(valStr, ",") {
		valStr = strings.ReplaceAll(valStr, ",", ".")
	}

	val, _ := strconv.ParseFloat(valStr, 64)
	return val
}
