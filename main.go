package main

import (
	"log"
	"net/http"
	"os"

	"github.com/doITmagic/speak4me/pkg/api"
	"github.com/doITmagic/speak4me/pkg/config"
	"github.com/doITmagic/speak4me/pkg/db"
	"github.com/doITmagic/speak4me/pkg/tts"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/f5tts"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/mock"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/xtts"
)

func main() {
	// 0. Încărcăm configurația
	cfg := config.LoadConfig("config.json")

	// 1. Inițializăm Baza de date
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Eroare la pornirea bazei de date: %v", err)
	}
	defer database.Close()
	log.Println("✅ Baza de date SQLite inițializată.")

	// 2. Inițializăm Registry-ul de modele TTS
	registry := tts.NewRegistry()

	// 3. Înregistrăm Adaptorul Mock (fallback / teste UI)
	mockAdapter := mock.NewMockTTS()
	_ = mockAdapter.Init(map[string]interface{}{"delay_ms": 300})
	registry.RegisterProvider("mock_v1", mockAdapter)
	log.Println("✅ Adaptor Mock înregistrat.")

	// 4. Înregistrăm Adaptorul F5-TTS
	if cfg.F5TTS.Enabled {
		os.MkdirAll(cfg.F5TTS.VoicesDir, os.ModePerm)
		os.MkdirAll(cfg.F5TTS.ModelsDir, os.ModePerm)

		f5Adapter := f5tts.NewF5TTSAdapter()
		err = f5Adapter.Init(map[string]interface{}{
			"models_dir": cfg.F5TTS.ModelsDir,
			"voices_dir": cfg.F5TTS.VoicesDir,
		})
		if err != nil {
			log.Printf("⚠️  F5-TTS: %v", err)
			log.Println("   Poți rula: bash sidecar/export_f5_onnx.sh pentru a exporta modelul.")
			registry.RegisterProvider(f5tts.ModelID, f5Adapter)
		} else {
			registry.RegisterProvider(f5tts.ModelID, f5Adapter)
			log.Println("✅ Adaptor F5-TTS Romanian (ONNX) înregistrat.")
		}
	}

	// 5. Înregistrăm Adaptorul XTTS (Sidecar HTTP)
	if cfg.XTTS.Enabled {
		os.MkdirAll(cfg.XTTS.VoicesDir, os.ModePerm)
		
		xttsAdapter := xtts.NewXTTSAdapter()
		_ = xttsAdapter.Init(map[string]interface{}{
			"server_url": cfg.XTTS.ServerURL,
			"voices_dir": cfg.XTTS.VoicesDir,
		})
		registry.RegisterProvider(xtts.ModelID, xttsAdapter)
		log.Println("✅ Adaptor XTTS-v2 (Sidecar) înregistrat.")
	}

	// 6. Pornim Serverul HTTP
	server := api.NewServer(registry, database)

	log.Printf("🚀 Serverul Speak4Me pornește pe http://localhost%s", cfg.ServerPort)
	log.Println("   GET  /api/v1/voices    → Lista tuturor vocilor disponibile")
	log.Println("   POST /api/v1/synthesize → Sintetizare text în audio")
	if err := http.ListenAndServe(cfg.ServerPort, server.Router); err != nil {
		log.Fatalf("Server oprit: %v", err)
	}
}
