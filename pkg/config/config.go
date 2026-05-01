package config

import (
	"encoding/json"
	"log"
	"os"
)

// AppConfig definește structura fișierului de configurare
type AppConfig struct {
	ServerPort string    `json:"server_port"`
	DBPath     string    `json:"db_path"`
	F5TTS      TTSConfig `json:"f5_tts"`
	XTTS       TTSConfig `json:"xtts"`
}

// TTSConfig definește setările specifice pentru un model TTS
type TTSConfig struct {
	Enabled   bool   `json:"enabled"`
	ServerURL string `json:"server_url,omitempty"`
	ModelsDir string `json:"models_dir,omitempty"`
	VoicesDir string `json:"voices_dir,omitempty"`
}

// LoadConfig citește setările din config.json. Dacă nu există, îl creează cu valori default.
func LoadConfig(path string) *AppConfig {
	cfg := &AppConfig{
		ServerPort: ":8020",
		DBPath:     "speak4me.db",
		F5TTS: TTSConfig{
			Enabled:   true,
			ModelsDir: "./models/f5-tts-romanian",
			VoicesDir: "./voices/f5",
		},
		XTTS: TTSConfig{
			Enabled:   true,
			ServerURL: "http://127.0.0.1:8021",
			VoicesDir: "./voices/xtts",
		},
	}

	file, err := os.ReadFile(path)
	if err == nil {
		err = json.Unmarshal(file, cfg)
		if err != nil {
			log.Printf("⚠️  Eroare la parsarea %s: %v. Se folosesc valorile default.", path, err)
		} else {
			log.Printf("✅ Configurație încărcată cu succes din %s", path)
		}
	} else {
		log.Printf("ℹ️  Fişierul %s nu a fost găsit. Generăm config implicit.", path)
		bytes, _ := json.MarshalIndent(cfg, "", "  ")
		os.WriteFile(path, bytes, 0644)
	}

	return cfg
}
