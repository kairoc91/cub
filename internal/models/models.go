package models

import "time"

// CUBEntry representa uma entrada do histórico do CUB
type CUBEntry struct {
	AnoMes         string  `json:"ano_mes"`
	MesAno         string  `json:"mes_ano"`
	Valor          float64 `json:"valor"`
	ValorFormatado string  `json:"valor_formatado"`
}

// SimulacaoInput contém os dados informados pelo usuário no formulário
type SimulacaoInput struct {
	Titulo             string  `json:"titulo"`
	CUBBase            string  `json:"cub_base"`
	ValorSinal         float64 `json:"valor_sinal"`
	ValorEntrada       float64 `json:"valor_entrada"`
	TipoParcela        string  `json:"tipo_parcela"`
	QtdParcelas        int     `json:"qtd_parcelas"`
	PrimeiroVencimento string  `json:"primeiro_vencimento"` // YYYY-MM-DD
	ValorParcela       float64 `json:"valor_parcela"`
	ValorChaves        float64 `json:"valor_chaves"`
}

// MetricaItem consolida valores financeiros, percentual e equivalência CUB
type MetricaItem struct {
	Titulo              string  `json:"titulo"`
	Valor               float64 `json:"valor"`
	ValorFormatado      string  `json:"valor_formatado"`
	Percentual          float64 `json:"percentual"`
	PercentualFormatado string  `json:"percentual_formatado"`
	EquivCUB            float64 `json:"equiv_cub"`
	EquivCUBFormatado   string  `json:"equiv_cub_formatado"`
}

// CronogramaItem representa cada linha da tabela de fluxo de pagamentos
type CronogramaItem struct {
	Numero                      int     `json:"numero"`
	Descricao                   string  `json:"descricao"`
	Vencimento                  string  `json:"vencimento"` // DD/MM/AAAA
	Periodicidade               string  `json:"periodicidade"`
	Valor                       float64 `json:"valor"`
	ValorFormatado              string  `json:"valor_formatado"`
	EquivCUB                    float64 `json:"equiv_cub"`
	EquivCUBFormatado           string  `json:"equiv_cub_formatado"`
	ValorCUBAtual               float64 `json:"valor_cub_atual"`
	ValorCUBAtualFormatado      string  `json:"valor_cub_atual_formatado"`
	EquivCUBAcumulado           float64 `json:"equiv_cub_acumulado"`
	AcumuladoCUBAtual           float64 `json:"acumulado_cub_atual"`
	AcumuladoCUBAtualFormatado  string  `json:"acumulado_cub_atual_formatado"`
}

// SimulacaoResultado representa todo o balanço financeiro calculado
type SimulacaoResultado struct {
	TemValores                 bool             `json:"tem_valores"`
	TotalGeral                 float64          `json:"total_geral"`
	TotalGeralFormatado        string           `json:"total_geral_formatado"`
	TotalCUBs                  float64          `json:"total_cubs"`
	TotalCUBsFormatado         string           `json:"total_cubs_formatado"`
	TotalProjetadoAtual        float64          `json:"total_projetado_atual"`
	TotalProjetadoAtualFormatado string         `json:"total_projetado_atual_formatado"`
	CUBBaseValor               float64          `json:"cub_base_valor"`
	CUBBaseFormatado           string           `json:"cub_base_formatado"`
	CUBBaseMesAno              string           `json:"cub_base_mes_ano"`
	CUBAtualValor              float64          `json:"cub_atual_valor"`
	CUBAtualFormatado          string           `json:"cub_atual_formatado"`
	CUBAtualMesAno             string           `json:"cub_atual_mes_ano"`
	Sinal                      MetricaItem      `json:"sinal"`
	Entrada                    MetricaItem      `json:"entrada"`
	Parcelas                   MetricaItem      `json:"parcelas"`
	Chaves                     MetricaItem      `json:"chaves"`
	PctSinal                   float64          `json:"pct_sinal"`
	PctEntrada                 float64          `json:"pct_entrada"`
	PctParcelas                float64          `json:"pct_parcelas"`
	PctChaves                  float64          `json:"pct_chaves"`
	Cronograma                 []CronogramaItem `json:"cronograma"`
	TotalItensCronograma       int              `json:"total_itens_cronograma"`
}

// PropostaSalva armazena uma proposta persistida
type PropostaSalva struct {
	ID                string              `json:"id"`
	CriadaEm          time.Time           `json:"criada_em"`
	CriadaEmFormatada string              `json:"criada_em_formatada"`
	Input             SimulacaoInput      `json:"input"`
	Resultado         SimulacaoResultado  `json:"resultado"`
}

// PageData é o envelope enviado para a renderização inicial do HTML
type PageData struct {
	CUBHistorico []CUBEntry          `json:"cub_historico"`
	CUBMaisRecente CUBEntry          `json:"cub_mais_recente"`
	Input        SimulacaoInput      `json:"input"`
	Resultado    SimulacaoResultado  `json:"resultado"`
	Propostas    []PropostaSalva     `json:"propostas"`
	ToastMensagem string             `json:"toast_mensagem"`
}
