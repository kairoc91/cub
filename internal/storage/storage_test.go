package storage

import (
	"path/filepath"
	"testing"

	"cub/internal/models"
)

func TestStorage_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "propostas_test.json")

	store, err := NewStorage(dbPath)
	if err != nil {
		t.Fatalf("falha ao criar storage: %v", err)
	}

	if len(store.List()) != 0 {
		t.Errorf("esperava lista vazia inicialmente")
	}

	input := models.SimulacaoInput{
		Titulo:     "Proposta Apto 101",
		ValorSinal: 50000,
	}
	res := models.SimulacaoResultado{
		TotalGeral: 50000,
	}

	salva, err := store.Save(input, res)
	if err != nil {
		t.Fatalf("falha ao salvar proposta: %v", err)
	}

	if salva.ID == "" {
		t.Errorf("ID gerado não deve ser vazio")
	}

	items := store.List()
	if len(items) != 1 {
		t.Fatalf("esperava 1 item na lista, obteve %d", len(items))
	}
	if items[0].Input.Titulo != "Proposta Apto 101" {
		t.Errorf("título incorreto: %s", items[0].Input.Titulo)
	}

	rec, found := store.Get(salva.ID)
	if !found || rec.ID != salva.ID {
		t.Errorf("não encontrou proposta por ID")
	}

	if err := store.Delete(salva.ID); err != nil {
		t.Fatalf("falha ao deletar proposta: %v", err)
	}
	if len(store.List()) != 0 {
		t.Errorf("esperava lista vazia após delete")
	}
}
