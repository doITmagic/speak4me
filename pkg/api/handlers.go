package api

import (
	"encoding/json"
	"net/http"

	"github.com/doITmagic/speak4me/pkg/tts"
)

// handleGetVoices returnează lista de voci disponibile din sistem (format JSON)
func (s *APIServer) handleGetVoices() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		voices, err := s.Registry.ListAllVoices()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(voices)
	}
}

// handleSynthesize primește textul și sintetizează audio-ul returnând byte array-ul
func (s *APIServer) handleSynthesize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tts.SynthesisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		if req.VoiceID == "" || req.Text == "" {
			http.Error(w, "VoiceID and Text are required fields", http.StatusBadRequest)
			return
		}

		// Obținem providerul corect deși interfața noastră simplificată de la `Manager` (de făcut mai târziu un wrapper)
		// Pentru moment iterăm prin provideri (sau modificăm Manager-ul) pentru a găsi vocea. 
		// Din specificații, ModelRegistry ne ajută. Să simplificăm și să presupunem că providerul primește ID-ul.
		// Din design, fiecare voice are ModelID. Întâi trebuie să găsim ce model deține acea voce.
		
		allVoices, _ := s.Registry.ListAllVoices()
		var targetModelID string
		for _, v := range allVoices {
			if v.ID == req.VoiceID {
				targetModelID = v.ModelID
				break
			}
		}

		if targetModelID == "" {
			http.Error(w, "Voice not found", http.StatusNotFound)
			return
		}

		provider, err := s.Registry.GetProvider(targetModelID)
		if err != nil {
			http.Error(w, "Provider internal error", http.StatusInternalServerError)
			return
		}

		audioData, err := provider.Synthesize(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioData)
	}
}
