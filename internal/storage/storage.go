package storage

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"cub/internal/models"
)

// Storage gerencia a persistência de propostas de forma concorrente e atômica
type Storage struct {
	mu       sync.RWMutex
	filePath string
	items    []models.PropostaSalva
}

// NewStorage inicializa o armazenamento baseado em arquivo JSON local
func NewStorage(filePath string) (*Storage, error) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("falha ao criar diretório de dados (%s): %w", dir, err)
	}

	s := &Storage{
		filePath: filePath,
		items:    make([]models.PropostaSalva, 0),
	}

	if _, err := os.Stat(filePath); err == nil {
		data, err := os.ReadFile(filePath)
		if err == nil && len(data) > 0 {
			var loaded []models.PropostaSalva
			if err := json.Unmarshal(data, &loaded); err == nil {
				s.items = loaded
			}
		}
	}

	return s, nil
}

// List retorna todas as propostas salvas ordenadas da mais recente para a mais antiga
func (s *Storage) List() []models.PropostaSalva {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.PropostaSalva, len(s.items))
	copy(result, s.items)
	return result
}

// Get obtém uma proposta por ID
func (s *Storage) Get(id string) (models.PropostaSalva, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.items {
		if item.ID == id {
			return item, true
		}
	}
	return models.PropostaSalva{}, false
}

// Save adiciona ou atualiza uma proposta
func (s *Storage) Save(input models.SimulacaoInput, res models.SimulacaoResultado) (models.PropostaSalva, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idBytes := make([]byte, 6)
	_, _ = rand.Read(idBytes)
	id := fmt.Sprintf("%x-%d", idBytes, time.Now().Unix())

	agora := time.Now()
	proposta := models.PropostaSalva{
		ID:                id,
		CriadaEm:          agora,
		CriadaEmFormatada: agora.Format("02/01/2006 às 15:04"),
		Input:             input,
		Resultado:         res,
	}

	// Insere no início da lista para aparecer primeiro
	s.items = append([]models.PropostaSalva{proposta}, s.items...)

	if err := s.saveToFileLocked(); err != nil {
		return models.PropostaSalva{}, err
	}

	return proposta, nil
}

// Delete remove uma proposta pelo ID
func (s *Storage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var novos []models.PropostaSalva
	for _, item := range s.items {
		if item.ID != id {
			novos = append(novos, item)
		}
	}
	s.items = novos
	return s.saveToFileLocked()
}

// ClearAll remove todas as propostas salvas
func (s *Storage) ClearAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make([]models.PropostaSalva, 0)
	return s.saveToFileLocked()
}

func (s *Storage) saveToFileLocked() error {
	data, err := json.MarshalIndent(s.items, "", "  ")
	if err != nil {
		return fmt.Errorf("erro ao serializar propostas: %w", err)
	}

	// Escrita atômica via arquivo temporário para evitar corrupção em falhas elétricas/processos
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return fmt.Errorf("erro ao gravar arquivo temporário de propostas: %w", err)
	}

	if err := os.Rename(tmpFile, s.filePath); err != nil {
		return fmt.Errorf("erro ao atualizar arquivo de propostas: %w", err)
	}

	return nil
}
