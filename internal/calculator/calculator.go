package calculator

import (
	"fmt"
	"math"
	"strings"
	"time"

	"cub/internal/cub"
	"cub/internal/models"
)

// Calculator implementa a lógica do motor financeiro de propostas
type Calculator struct {
	repo *cub.Repository
}

// NewCalculator cria uma nova instância do calculador
func NewCalculator(repo *cub.Repository) *Calculator {
	return &Calculator{repo: repo}
}

// Calcular processa os parâmetros de entrada e gera o resultado completo
func (c *Calculator) Calcular(input models.SimulacaoInput) models.SimulacaoResultado {
	cubAtual := c.repo.GetMaisRecente()

	cubBase := cubAtual
	if input.CUBBase != "" {
		if found, ok := c.repo.FindByAnoMes(input.CUBBase); ok {
			cubBase = found
		}
	}

	subtotalParcelas := float64(input.QtdParcelas) * input.ValorParcela
	totalGeral := input.ValorSinal + input.ValorEntrada + subtotalParcelas + input.ValorChaves

	resultado := models.SimulacaoResultado{
		TemValores:        totalGeral > 0,
		TotalGeral:        totalGeral,
		TotalGeralFormatado: cub.FormatCurrency(totalGeral),
		CUBBaseValor:      cubBase.Valor,
		CUBBaseFormatado:  cubBase.ValorFormatado,
		CUBBaseMesAno:     cubBase.MesAno,
		CUBAtualValor:     cubAtual.Valor,
		CUBAtualFormatado: cubAtual.ValorFormatado,
		CUBAtualMesAno:    cubAtual.MesAno,
	}

	if totalGeral <= 0 || cubBase.Valor <= 0 {
		return resultado
	}

	resultado.TotalCUBs = totalGeral / cubBase.Valor
	resultado.TotalCUBsFormatado = formatCUB(resultado.TotalCUBs)
	resultado.TotalProjetadoAtual = resultado.TotalCUBs * cubAtual.Valor
	resultado.TotalProjetadoAtualFormatado = cub.FormatCurrency(resultado.TotalProjetadoAtual)

	// Percentuais e métricas
	pctSinal := (input.ValorSinal / totalGeral) * 100
	cubSinal := input.ValorSinal / cubBase.Valor
	resultado.Sinal = models.MetricaItem{
		Titulo:              "Sinal",
		Valor:               input.ValorSinal,
		ValorFormatado:      cub.FormatCurrency(input.ValorSinal),
		Percentual:          pctSinal,
		PercentualFormatado: formatPercent(pctSinal),
		EquivCUB:            cubSinal,
		EquivCUBFormatado:   formatCUB(cubSinal),
	}
	resultado.PctSinal = pctSinal

	pctEntrada := (input.ValorEntrada / totalGeral) * 100
	cubEntrada := input.ValorEntrada / cubBase.Valor
	resultado.Entrada = models.MetricaItem{
		Titulo:              "Entrada",
		Valor:               input.ValorEntrada,
		ValorFormatado:      cub.FormatCurrency(input.ValorEntrada),
		Percentual:          pctEntrada,
		PercentualFormatado: formatPercent(pctEntrada),
		EquivCUB:            cubEntrada,
		EquivCUBFormatado:   formatCUB(cubEntrada),
	}
	resultado.PctEntrada = pctEntrada

	pctParcelas := (subtotalParcelas / totalGeral) * 100
	cubParcelas := subtotalParcelas / cubBase.Valor
	resultado.Parcelas = models.MetricaItem{
		Titulo:              "Subtotal Parcelas",
		Valor:               subtotalParcelas,
		ValorFormatado:      cub.FormatCurrency(subtotalParcelas),
		Percentual:          pctParcelas,
		PercentualFormatado: formatPercent(pctParcelas),
		EquivCUB:            cubParcelas,
		EquivCUBFormatado:   formatCUB(cubParcelas),
	}
	resultado.PctParcelas = pctParcelas

	pctChaves := (input.ValorChaves / totalGeral) * 100
	cubChaves := input.ValorChaves / cubBase.Valor
	resultado.Chaves = models.MetricaItem{
		Titulo:              "Chaves",
		Valor:               input.ValorChaves,
		ValorFormatado:      cub.FormatCurrency(input.ValorChaves),
		Percentual:          pctChaves,
		PercentualFormatado: formatPercent(pctChaves),
		EquivCUB:            cubChaves,
		EquivCUBFormatado:   formatCUB(cubChaves),
	}
	resultado.PctChaves = pctChaves

	// Construção do Cronograma
	resultado.Cronograma = c.construirCronograma(input, cubBase.Valor, cubAtual.Valor)
	resultado.TotalItensCronograma = len(resultado.Cronograma)

	return resultado
}

func (c *Calculator) construirCronograma(input models.SimulacaoInput, cubBaseVal, cubAtualVal float64) []models.CronogramaItem {
	var cronograma []models.CronogramaItem
	itemNum := 1
	var cubAcumulado float64

	var dataSinal time.Time
	var dataEntrada time.Time
	var dataParcela time.Time

	if input.PrimeiroVencimento != "" {
		if t, err := time.Parse("2006-01-02", input.PrimeiroVencimento); err == nil {
			dataParcela = t
			// Sinal ocorre no mês de assinatura/contratação (1 mês antes do 1º vencimento da parcela)
			dataSinal = dataParcela.AddDate(0, -1, 0)
			// Entrada/Reforço ocorre 15 dias após o Sinal
			dataEntrada = dataSinal.AddDate(0, 0, 15)
		}
	}

	if dataParcela.IsZero() {
		hoje := time.Now()
		dataSinal = hoje
		dataEntrada = hoje.AddDate(0, 0, 15)
		dataParcela = hoje.AddDate(0, 1, 0)
	}

	// 1. Sinal
	if input.ValorSinal > 0 {
		equivCUB := input.ValorSinal / cubBaseVal
		valorCUBAtual := equivCUB * cubAtualVal
		cubAcumulado += equivCUB

		cItem := models.CronogramaItem{
			Numero:                     itemNum,
			Descricao:                  "Sinal (Ato)",
			Vencimento:                 dataSinal.Format("02/01/2006"),
			Periodicidade:              "Ato",
			Valor:                      input.ValorSinal,
			ValorFormatado:             cub.FormatCurrency(input.ValorSinal),
			EquivCUB:                   equivCUB,
			EquivCUBFormatado:          formatCUB(equivCUB),
			ValorCUBAtual:              valorCUBAtual,
			ValorCUBAtualFormatado:     cub.FormatCurrency(valorCUBAtual),
			EquivCUBAcumulado:          cubAcumulado,
			AcumuladoCUBAtual:          cubAcumulado * cubAtualVal,
			AcumuladoCUBAtualFormatado: cub.FormatCurrency(cubAcumulado * cubAtualVal),
		}

		cronograma = append(cronograma, cItem)
		itemNum++
	}

	// 2. Entrada
	if input.ValorEntrada > 0 {
		equivCUB := input.ValorEntrada / cubBaseVal
		valorCUBAtual := equivCUB * cubAtualVal
		cubAcumulado += equivCUB

		cItem := models.CronogramaItem{
			Numero:                     itemNum,
			Descricao:                  "Reforço de Entrada",
			Vencimento:                 dataEntrada.Format("02/01/2006"),
			Periodicidade:              "Entrada",
			Valor:                      input.ValorEntrada,
			ValorFormatado:             cub.FormatCurrency(input.ValorEntrada),
			EquivCUB:                   equivCUB,
			EquivCUBFormatado:          formatCUB(equivCUB),
			ValorCUBAtual:              valorCUBAtual,
			ValorCUBAtualFormatado:     cub.FormatCurrency(valorCUBAtual),
			EquivCUBAcumulado:          cubAcumulado,
			AcumuladoCUBAtual:          cubAcumulado * cubAtualVal,
			AcumuladoCUBAtualFormatado: cub.FormatCurrency(cubAcumulado * cubAtualVal),
		}

		cronograma = append(cronograma, cItem)
		itemNum++
	}

	// 3. Parcelas Periódicas
	mesesSalto := 1
	rotuloPeriodo := "Mensal"
	switch strings.ToLower(input.TipoParcela) {
	case "trimestral":
		mesesSalto = 3
		rotuloPeriodo = "Trimestral"
	case "semestral":
		mesesSalto = 6
		rotuloPeriodo = "Semestral"
	case "anual":
		mesesSalto = 12
		rotuloPeriodo = "Anual"
	default:
		mesesSalto = 1
		rotuloPeriodo = "Mensal"
	}

	for i := 1; i <= input.QtdParcelas && input.ValorParcela > 0; i++ {
		vencimento := dataParcela.AddDate(0, (i-1)*mesesSalto, 0)
		equivCUB := input.ValorParcela / cubBaseVal
		valorCUBAtual := equivCUB * cubAtualVal
		cubAcumulado += equivCUB

		cItem := models.CronogramaItem{
			Numero:                     itemNum,
			Descricao:                  fmt.Sprintf("Parcela %02d/%02d", i, input.QtdParcelas),
			Vencimento:                 vencimento.Format("02/01/2006"),
			Periodicidade:              rotuloPeriodo,
			Valor:                      input.ValorParcela,
			ValorFormatado:             cub.FormatCurrency(input.ValorParcela),
			EquivCUB:                   equivCUB,
			EquivCUBFormatado:          formatCUB(equivCUB),
			ValorCUBAtual:              valorCUBAtual,
			ValorCUBAtualFormatado:     cub.FormatCurrency(valorCUBAtual),
			EquivCUBAcumulado:          cubAcumulado,
			AcumuladoCUBAtual:          cubAcumulado * cubAtualVal,
			AcumuladoCUBAtualFormatado: cub.FormatCurrency(cubAcumulado * cubAtualVal),
		}

		cronograma = append(cronograma, cItem)
		itemNum++
	}

	// 4. Parcela nas Chaves
	if input.ValorChaves > 0 {
		vencimentoChaves := dataParcela.AddDate(0, input.QtdParcelas*mesesSalto, 0)
		equivCUB := input.ValorChaves / cubBaseVal
		valorCUBAtual := equivCUB * cubAtualVal
		cubAcumulado += equivCUB

		cItem := models.CronogramaItem{
			Numero:                     itemNum,
			Descricao:                  "Saldo nas Chaves",
			Vencimento:                 vencimentoChaves.Format("02/01/2006"),
			Periodicidade:              "Chaves / Entrega",
			Valor:                      input.ValorChaves,
			ValorFormatado:             cub.FormatCurrency(input.ValorChaves),
			EquivCUB:                   equivCUB,
			EquivCUBFormatado:          formatCUB(equivCUB),
			ValorCUBAtual:              valorCUBAtual,
			ValorCUBAtualFormatado:     cub.FormatCurrency(valorCUBAtual),
			EquivCUBAcumulado:          cubAcumulado,
			AcumuladoCUBAtual:          cubAcumulado * cubAtualVal,
			AcumuladoCUBAtualFormatado: cub.FormatCurrency(cubAcumulado * cubAtualVal),
		}

		cronograma = append(cronograma, cItem)
	}

	return cronograma
}

func formatCUB(val float64) string {
	val = math.Round(val*10000) / 10000
	str := fmt.Sprintf("%.4f", val)
	return strings.ReplaceAll(str, ".", ",")
}

func formatPercent(val float64) string {
	val = math.Round(val*10) / 10
	str := fmt.Sprintf("%.1f%%", val)
	return strings.ReplaceAll(str, ".", ",")
}
