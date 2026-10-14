package cub

import (
	"strings"
	"testing"
)

func TestNewRepositoryFromReader(t *testing.T) {
	csvData := `ano_mes;mes_ano;valor;valor_formatado
2026-08;08/2026;3151.2400;"R$ 3.151,2400"
2026-09;09/2026;3158.8800;"R$ 3.158,8800"
2026-10;10/2026;3171.9600;"R$ 3.171,9600"
`
	repo, err := NewRepositoryFromReader(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("falha ao carregar repositório: %v", err)
	}

	maisRecente := repo.GetMaisRecente()
	if maisRecente.AnoMes != "2026-10" {
		t.Errorf("esperava mais recente 2026-10, obteve %s", maisRecente.AnoMes)
	}
	if maisRecente.Valor != 3171.96 {
		t.Errorf("esperava valor 3171.96, obteve %f", maisRecente.Valor)
	}

	entry, found := repo.FindByAnoMes("2026-08")
	if !found {
		t.Errorf("entrada 2026-08 não encontrada")
	}
	if entry.MesAno != "08/2026" {
		t.Errorf("esperava mes_ano 08/2026, obteve %s", entry.MesAno)
	}
}

func TestFormatCurrency(t *testing.T) {
	tests := []struct {
		input    float64
		expected string
	}{
		{0, "R$ 0,00"},
		{1250.50, "R$ 1.250,50"},
		{1000000.00, "R$ 1.000.000,00"},
		{3171.96, "R$ 3.171,96"},
	}

	for _, tt := range tests {
		got := FormatCurrency(tt.input)
		if got != tt.expected {
			t.Errorf("FormatCurrency(%f) = %s; esperava %s", tt.input, got, tt.expected)
		}
	}
}
