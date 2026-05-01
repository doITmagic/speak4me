package tts

import (
	"errors"
	"sync"
)

var (
	ErrModelNotFound = errors.New("model provider not found")
	ErrProviderInit  = errors.New("failed to init provider")
)

// ModelRegistry orchestrează și păstrează legătura cu instanțele diferitelor modele TTS
type ModelRegistry struct {
	providers map[string]TTSProvider
	mu        sync.RWMutex
}

// NewRegistry instanțiază un registry go
func NewRegistry() *ModelRegistry {
	return &ModelRegistry{
		providers: make(map[string]TTSProvider),
	}
}

// RegisterProvider adaugă un adaptor TTS în registru
func (r *ModelRegistry) RegisterProvider(modelID string, provider TTSProvider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.providers[modelID] = provider
	return nil
}

// GetProvider recuperează un adaptor TTS pe baza ID-ului modelului
func (r *ModelRegistry) GetProvider(modelID string) (TTSProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[modelID]
	if !exists {
		return nil, ErrModelNotFound
	}
	return provider, nil
}

// ListAllVoices adună toate profilurile vocale de la toate provider-urile active
func (r *ModelRegistry) ListAllVoices() ([]VoiceProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var allVoices []VoiceProfile

	for _, provider := range r.providers {
		voices, err := provider.GetAvailableVoices()
		if err != nil {
			// Poți alege să loghezi și să sari modelul respectiv, sau să dai return
			continue
		}
		allVoices = append(allVoices, voices...)
	}

	return allVoices, nil
}
