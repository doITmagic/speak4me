package api

import (
	"encoding/json"
	"net/http"

	"github.com/doITmagic/speak4me/pkg/audio"
	"github.com/doITmagic/speak4me/pkg/tts"
)

// handleStreamSynthesize folosește Server-Sent Events (SSE) sau Transfer-Encoding: chunked 
// pentru a trimite raw audio pe măsură ce este generat de RollingBuffer.
func (s *APIServer) handleStreamSynthesize() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
			return
		}

		var req tts.SynthesisRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if req.VoiceID == "" || req.Text == "" {
			http.Error(w, "VoiceID and Text required", http.StatusBadRequest)
			return
		}

		// Identificarea modelului
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

		// Inițializarea Chunking-ului și Buffer-ului
		sentences := audio.SplitTextIntoSentences(req.Text)
		inputChan := make(chan string, len(sentences))
		for _, s := range sentences {
			inputChan <- s
		}
		close(inputChan)

		buffer := audio.NewRollingBuffer(provider, req.VoiceID)
		audioChan, errChan := buffer.ProcessStream(inputChan)

		w.Header().Set("Content-Type", "audio/wav") // sau application/octet-stream pentru un flux continuu PCM raw
		w.Header().Set("Transfer-Encoding", "chunked")
		w.Header().Set("Connection", "keep-alive")

		// Consumăm din buffer și trimitem "pe țeavă" spre frontend/client
		for {
			select {
			case err := <-errChan:
				if err != nil {
					// Dacă apare o eroare la mijloc, întrerupem transmisia
					return 
				}
			case audioData, ok := <-audioChan:
				if !ok {
					// Stream-ul s-a terminat cu succes
					return
				}
				// Trimitem bytes-ul audio
				_, writeErr := w.Write(audioData)
				if writeErr != nil {
					return // Clientul s-a deconectat prematur
				}
				flusher.Flush() // Forțăm livrarea către rețea
			case <-r.Context().Done():
				// Client-ul HTTP a închis conexiunea
				return
			}
		}
	}
}
