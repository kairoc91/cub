package calculator

import (
	"strings"
	"testing"

	"cub/internal/cub"
	"cub/internal/models"
)

func setupTestRepo(t *testing.T) *cub.Repository {
	csvData := `ano_mes;mes_ano;valor;valor_formatado
2026-09;09/2026;3000.0000;"R$ 3.000,0000"
2026-10;10/2026;3200.0000;"R$ 3.200,0000"
`
	repo, err := cub.NewRepositoryFromReader(strings.NewReader(csvData))
	if err != nil {
		t.Fatalf("falha ao criar repo de teste: %v", err)
	}
	return repo
}

func TestCalcular_CenarioBasico(t *testing.T) {
	repo := setupTestRepo(t)
	calc := NewCalculator(repo)

	input := models.SimulacaoInput{
		Titulo:             "Proposta Teste",
		CUBBase:            "2026-09", // Valor R$ 3.000,00
		ValorSinal:         30000.0,
		ValorEntrada:       60000.0,
		TipoParcela:        "mensal",
		QtdParcelas:        10,
		ValorParcela:       15000.0, // Subtotal = 150.000
		ValorChaves:        60000.0,
		PrimeiroVencimento: "2026-11-10",
	}

	res := calc.Calcular(input)

	// Total geral = 30k + 60k + 150k + 60k = 300k
	if res.TotalGeral != 300000.0 {
		t.Errorf("esperava TotalGeral 300000.0, obteve %f", res.TotalGeral)
	}

	// Total CUBs = 300.000 / 3.000 = 100 CUBs
	if res.TotalCUBs != 100.0 {
		t.Errorf("esperava TotalCUBs 100.0, obteve %f", res.TotalCUBs)
	}

	// Sinal: 30k de 300k = 10%
	if res.PctSinal != 10.0 {
		t.Errorf("esperava PctSinal 10%%, obteve %f", res.PctSinal)
	}
	if res.Sinal.EquivCUB != 10.0 {
		t.Errorf("esperava Sinal CUB 10, obteve %f", res.Sinal.EquivCUB)
	}

	// Cronograma: 1 sinal + 1 entrada + 10 parcelas + 1 chaves = 13 itens
	if len(res.Cronograma) != 13 {
		t.Errorf("esperava 13 itens no cronograma, obteve %d", len(res.Cronograma))
	}

	// Verifica cálculo individual no CUB atual e datas em ordem cronológica
	sinalItem := res.Cronograma[0]
	if sinalItem.Vencimento != "10/10/2026" {
		t.Errorf("esperava vencimento do sinal 10/10/2026, obteve %s", sinalItem.Vencimento)
	}
	if sinalItem.ValorCUBAtual != 32000.0 {
		t.Errorf("esperava ValorCUBAtual do sinal 32000.0, obteve %f", sinalItem.ValorCUBAtual)
	}

	entradaItem := res.Cronograma[1]
	if entradaItem.Vencimento != "25/10/2026" {
		t.Errorf("esperava vencimento da entrada 25/10/2026, obteve %s", entradaItem.Vencimento)
	}
	if entradaItem.ValorCUBAtual != 64000.0 {
		t.Errorf("esperava ValorCUBAtual da entrada 64000.0, obteve %f", entradaItem.ValorCUBAtual)
	}

	parcela1 := res.Cronograma[2]
	if parcela1.Vencimento != "10/11/2026" {
		t.Errorf("esperava vencimento da parcela 1 10/11/2026, obteve %s", parcela1.Vencimento)
	}
	if parcela1.ValorCUBAtual != 16000.0 {
		t.Errorf("esperava ValorCUBAtual da parcela 1 16000.0, obteve %f", parcela1.ValorCUBAtual)
	}

	// Total Projetado Atual = 100 CUBs * 3200 = 320.000
	if res.TotalProjetadoAtual != 320000.0 {
		t.Errorf("esperava TotalProjetadoAtual 320000.0, obteve %f", res.TotalProjetadoAtual)
	}

	// Último item (Chaves): CUB acumulado deve ser 100 CUBs
	ultimoItem := res.Cronograma[len(res.Cronograma)-1]
	if ultimoItem.EquivCUBAcumulado != 100.0 {
		t.Errorf("esperava CUB acumulado 100.0 no final, obteve %f", ultimoItem.EquivCUBAcumulado)
	}
	// Acumulado pelo CUB atual (3.200) = 100 * 3200 = 320.000
	if ultimoItem.AcumuladoCUBAtual != 320000.0 {
		t.Errorf("esperava AcumuladoCUBAtual 320000.0, obteve %f", ultimoItem.AcumuladoCUBAtual)
	}
}

func TestCalcular_Vazio(t *testing.T) {
	repo := setupTestRepo(t)
	calc := NewCalculator(repo)

	input := models.SimulacaoInput{}
	res := calc.Calcular(input)

	if res.TemValores {
		t.Errorf("esperava TemValores = false para input vazio")
	}
	if res.TotalGeral != 0 {
		t.Errorf("esperava TotalGeral = 0")
	}
}
