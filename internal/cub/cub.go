package cub

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"cub/internal/models"
)

// Repository gerencia o histórico de índices CUB
type Repository struct {
	entries []models.CUBEntry
	byAnoMes map[string]models.CUBEntry
}

// NewRepositoryFromFile carrega e indexa o arquivo CSV do CUB histórico
func NewRepositoryFromFile(filepath string) (*Repository, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo CSV do CUB (%s): %w", filepath, err)
	}
	defer file.Close()

	return NewRepositoryFromReader(file)
}

// NewRepositoryFromReader lê dados de qualquer io.Reader (útil para testes unitários)
func NewRepositoryFromReader(r io.Reader) (*Repository, error) {
	reader := csv.NewReader(r)
	reader.Comma = ';'
	reader.LazyQuotes = true

	// Lê cabeçalho
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("erro ao ler cabeçalho do CSV: %w", err)
	}

	colAnoMes, colMesAno, colValor, colValorFormatado := -1, -1, -1, -1
	for i, col := range header {
		clean := strings.TrimSpace(strings.ToLower(col))
		switch clean {
		case "ano_mes":
			colAnoMes = i
		case "mes_ano":
			colMesAno = i
		case "valor":
			colValor = i
		case "valor_formatado":
			colValorFormatado = i
		}
	}

	if colAnoMes == -1 || colMesAno == -1 || colValor == -1 {
		return nil, fmt.Errorf("colunas obrigatórias não encontradas no CSV (ano_mes, mes_ano, valor)")
	}

	var entries []models.CUBEntry
	byAnoMes := make(map[string]models.CUBEntry)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("erro ao ler linha do CSV: %w", err)
		}

		anoMes := strings.TrimSpace(record[colAnoMes])
		mesAno := strings.TrimSpace(record[colMesAno])
		valorStr := strings.TrimSpace(record[colValor])
		valorStr = strings.ReplaceAll(valorStr, ",", ".")

		val, err := strconv.ParseFloat(valorStr, 64)
		if err != nil {
			continue
		}

		valorFormatado := ""
		if colValorFormatado != -1 && colValorFormatado < len(record) {
			valorFormatado = strings.Trim(strings.TrimSpace(record[colValorFormatado]), "\"")
		}
		if valorFormatado == "" {
			valorFormatado = FormatCurrency(val)
		}

		entry := models.CUBEntry{
			AnoMes:         anoMes,
			MesAno:         mesAno,
			Valor:          val,
			ValorFormatado: valorFormatado,
		}

		entries = append(entries, entry)
		byAnoMes[anoMes] = entry
	}

	// Ordena do mais recente para o mais antigo (descendente)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].AnoMes > entries[j].AnoMes
	})

	return &Repository{
		entries:  entries,
		byAnoMes: byAnoMes,
	}, nil
}

// GetHistorico retorna todas as entradas ordenadas da mais recente para a mais antiga
func (r *Repository) GetHistorico() []models.CUBEntry {
	result := make([]models.CUBEntry, len(r.entries))
	copy(result, r.entries)
	return result
}

// GetMaisRecente retorna o CUB mais atual disponível
func (r *Repository) GetMaisRecente() models.CUBEntry {
	if len(r.entries) == 0 {
		return models.CUBEntry{}
	}
	return r.entries[0]
}

// FindByAnoMes busca uma entrada específica por YYYY-MM
func (r *Repository) FindByAnoMes(anoMes string) (models.CUBEntry, bool) {
	entry, exists := r.byAnoMes[anoMes]
	return entry, exists
}

// FormatCurrency formata float para moeda brasileira (R$ 1.234,56)
func FormatCurrency(val float64) string {
	sign := ""
	if val < 0 {
		sign = "-"
		val = -val
	}
	intPart := int64(val)
	fracPart := int64((val-float64(intPart))*100 + 0.5)
	if fracPart >= 100 {
		intPart++
		fracPart = 0
	}

	intStr := strconv.FormatInt(intPart, 10)
	var formattedInt strings.Builder
	l := len(intStr)
	for i, c := range intStr {
		if i > 0 && (l-i)%3 == 0 {
			formattedInt.WriteString(".")
		}
		formattedInt.WriteRune(c)
	}

	return fmt.Sprintf("%sR$ %s,%02d", sign, formattedInt.String(), fracPart)
}
