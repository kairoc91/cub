package handlers

import (
	"html/template"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	appAssets "cub"
	"cub/internal/calculator"
	"cub/internal/cub"
	"cub/internal/models"
	"cub/internal/storage"
)

func setupTestServer(t *testing.T) *Server {
	csvData := `ano_mes;mes_ano;valor;valor_formatado
2026-09;09/2026;3000.0000;"R$ 3.000,0000"
2026-10;10/2026;3200.0000;"R$ 3.200,0000"
`
	repo, err := cub.NewRepositoryFromReader(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("falha ao criar repo cub: %v", err)
	}

	calc := calculator.NewCalculator(repo)
	tmpDir := t.TempDir()
	store, err := storage.NewStorage(filepath.Join(tmpDir, "propostas.json"))
	if err != nil {
		t.Fatalf("falha ao criar storage: %v", err)
	}

	srv, err := NewServer(repo, calc, store, "../../static")
	if err != nil {
		t.Fatalf("falha ao criar server: %v", err)
	}

	return srv
}

func TestHandlers_Index(t *testing.T) {
	srv := setupTestServer(t)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperava status 200, obteve %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "equivalência do CUB") && !strings.Contains(body, "Equivalência CUB") {
		t.Errorf("corpo da página não contém título esperado")
	}
	if !strings.Contains(body, "10/2026") {
		t.Errorf("CUB mais recente não exibido no HTML")
	}
}

func TestEmbeddedTemplates(t *testing.T) {
	tmpl, err := template.ParseFS(appAssets.TemplatesFS, "templates/*.html", "templates/partials/*.html")
	if err != nil {
		t.Fatalf("ParseFS error: %v", err)
	}
	for _, tpl := range tmpl.Templates() {
		t.Logf("template name: %q", tpl.Name())
	}
	data := models.PageData{
		Input: models.SimulacaoInput{
			ValorSinal:   55000,
			ValorEntrada: 110000,
			ValorParcela: 13750,
			ValorChaves:  55000,
		},
	}
	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, "index.html", data); err != nil {
		t.Fatalf("ExecuteTemplate error: %v", err)
	}
	t.Logf("Rendered %d bytes", buf.Len())
}

func TestHandlers_Calcular(t *testing.T) {
	srv := setupTestServer(t)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	form := url.Values{}
	form.Set("titulo", "Proposta Teste")
	form.Set("cub_base", "2026-10")
	form.Set("valor_sinal", "32000")
	form.Set("valor_entrada", "64000")
	form.Set("tipo_parcela", "mensal")
	form.Set("qtd_parcelas", "10")
	form.Set("valor_parcela", "16000")
	form.Set("valor_chaves", "32000")

	req := httptest.NewRequest("POST", "/api/calcular", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperava status 200, obteve %d", rec.Code)
	}

	body := rec.Body.String()
	// Total = 32k + 64k + 160k + 32k = 288k
	if !strings.Contains(body, "288.000,00") {
		t.Errorf("cálculo de total geral não encontrado na resposta HTML: %s", body)
	}
}

func TestHandlers_SalvarProposta(t *testing.T) {
	srv := setupTestServer(t)
	mux := http.NewServeMux()
	srv.RegisterRoutes(mux)

	form := url.Values{}
	form.Set("titulo", "Apto 502 Torre Sul")
	form.Set("cub_base", "2026-10")
	form.Set("valor_sinal", "50000")
	form.Set("qtd_parcelas", "12")
	form.Set("valor_parcela", "10000")

	req := httptest.NewRequest("POST", "/api/propostas", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("esperava status 200, obteve %d", rec.Code)
	}

	hxTrigger := rec.Header().Get("HX-Trigger")
	if !strings.Contains(hxTrigger, "showToast") {
		t.Errorf("HX-Trigger header esperado não encontrado: %s", hxTrigger)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Apto 502 Torre Sul") {
		t.Errorf("proposta salva não renderizada na lista: %s", body)
	}
}
