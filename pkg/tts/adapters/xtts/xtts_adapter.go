package xtts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/doITmagic/speak4me/pkg/tts"
)

const ModelID = "eduardem/xtts-v2-romanian"

// XTTSAdapter implementează TTSProvider pentru modelul XTTS-v2.
// Deoarece XTTS necesită inferență ML avansată, acest adaptor
// comunică printr-un API HTTP local cu un proces "sidecar" (ex: Python/Coqui TTS).
type XTTSAdapter struct {
	serverURL string
	voicesDir string
	client    *http.Client
}

// NewXTTSAdapter instanțiază adaptorul cu valori implicite
func NewXTTSAdapter() *XTTSAdapter {
	return &XTTSAdapter{
		serverURL: "http://127.0.0.1:8021", // Portul default pentru serverul sidecar XTTS
		voicesDir: "./voices/xtts",         // Directorul de unde citește wav-urile de clonare
		client:    &http.Client{},
	}
}

// Init aplică configurații opționale din map
func (a *XTTSAdapter) Init(config map[string]interface{}) error {
	if config != nil {
		if url, ok := config["server_url"].(string); ok && url != "" {
			a.serverURL = url
		}
		if dir, ok := config["voices_dir"].(string); ok && dir != "" {
			a.voicesDir = dir
		}
	}

	// Creăm directorul de voci dacă nu există pentru a preveni erori la scanare
	_ = os.MkdirAll(a.voicesDir, os.ModePerm)

	// Aici se poate adăuga o verificare (ex. un ping HTTP /health) către sidecar, 
	// dar îl vom menține decoupled pentru ca Go-ul să pornească chiar dacă sidecar-ul încă se încarcă.
	return nil
}

// GetAvailableVoices scanează directorul `voicesDir` după fișiere .wav
// XTTS folosește aceste fișiere pentru speaker conditioning (clonarea vocii).
func (a *XTTSAdapter) GetAvailableVoices() ([]tts.VoiceProfile, error) {
	var voices []tts.VoiceProfile

	files, err := os.ReadDir(a.voicesDir)
	if err != nil {
		// Returnăm array gol dacă directorul nu e accesibil
		return voices, nil 
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".wav") {
			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			
			voices = append(voices, tts.VoiceProfile{
				ID:       "xtts_ro_" + name,
				ModelID:  ModelID,
				Language: "ro-RO",
				Name:     strings.Title(name),
				Tags:     []string{"xtts", "romanian", "cloned"},
			})
		}
	}

	// Adăugăm o voce default fallback dacă folderul e gol
	if len(voices) == 0 {
		voices = append(voices, tts.VoiceProfile{
			ID:       "xtts_ro_default",
			ModelID:  ModelID,
			Language: "ro-RO",
			Name:     "XTTS Romanian Default",
			Tags:     []string{"xtts", "romanian", "default"},
		})
	}

	return voices, nil
}

// Synthesize trimite cererea către serverul sidecar XTTS și preia răspunsul audio
func (a *XTTSAdapter) Synthesize(req tts.SynthesisRequest) (tts.AudioStream, error) {
	// Preluăm numele de referință (fără prefixul xtts_ro_)
	speakerName := strings.TrimPrefix(req.VoiceID, "xtts_ro_")
	
	// Construim payload-ul pentru sidecar-ul Python
	payload := map[string]interface{}{
		"text":     req.Text,
		"speaker":  speakerName, // Sidecar-ul trebuie să caute referința wav a acestui speaker
		"language": "ro",        // XTTS v2 multi-lingual, fixăm pe română 
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("eroare json marshal: %w", err)
	}

	resp, err := a.client.Post(a.serverURL+"/api/tts", "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, fmt.Errorf("eroare conexiune cu sidecar XTTS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyErr, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("eroare de la modelul XTTS (status %d): %s", resp.StatusCode, string(bodyErr))
	}

	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("eroare citire audio: %w", err)
	}

	return tts.AudioStream(audioData), nil
}
