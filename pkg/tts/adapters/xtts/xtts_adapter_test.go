package xtts_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/doITmagic/speak4me/pkg/tts"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/xtts"
)

var _ = Describe("XttsAdapter", func() {
	var (
		adapter    *xtts.XTTSAdapter
		mockServer *httptest.Server
		testDir    string
	)

	BeforeEach(func() {
		// Mock pentru serverul de inferență Python/Sidecar XTTS
		mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/tts" {
				var payload map[string]interface{}
				_ = json.NewDecoder(r.Body).Decode(&payload)

				if payload["text"] == "" || payload["speaker"] == "" {
					http.Error(w, "Bad Request", http.StatusBadRequest)
					return
				}

				// Returnăm un fișier audio simulat "wav"
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("WAVE_AUDIO_BYTES_MOCK"))
				return
			}
			http.Error(w, "Not found", http.StatusNotFound)
		}))

		// Creăm un folder temporar pentru voci simulate
		testDir, _ = os.MkdirTemp("", "xtts_voices")
		_ = os.WriteFile(filepath.Join(testDir, "eduard.wav"), []byte("audio"), 0644)
		_ = os.WriteFile(filepath.Join(testDir, "maria.wav"), []byte("audio"), 0644)

		adapter = xtts.NewXTTSAdapter()
		err := adapter.Init(map[string]interface{}{
			"server_url": mockServer.URL,
			"voices_dir": testDir,
		})
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		mockServer.Close()
		os.RemoveAll(testDir)
	})

	Describe("Extragerea Vocilor de Referință (Speaker Conditioning)", func() {
		It("trebuie să găsească fișierele .wav și să le formateze corect", func() {
			voices, err := adapter.GetAvailableVoices()
			
			Expect(err).ToNot(HaveOccurred())
			Expect(len(voices)).To(Equal(2))

			var ids []string
			for _, v := range voices {
				ids = append(ids, v.ID)
			}
			Expect(ids).To(ContainElements("xtts_ro_eduard", "xtts_ro_maria"))
		})

		It("trebuie să ofere o voce de fallback dacă folderul e gol", func() {
			emptyDir, _ := os.MkdirTemp("", "xtts_empty")
			defer os.RemoveAll(emptyDir)

			emptyAdapter := xtts.NewXTTSAdapter()
			_ = emptyAdapter.Init(map[string]interface{}{"voices_dir": emptyDir})

			voices, _ := emptyAdapter.GetAvailableVoices()
			Expect(len(voices)).To(Equal(1))
			Expect(voices[0].ID).To(Equal("xtts_ro_default"))
		})
	})

	Describe("Sintetizare Audio (Inference)", func() {
		Context("când este chemat cu text valid", func() {
			It("trebuie să trimită HTTP corect la sidecar și să preia array-ul de bytes", func() {
				req := tts.SynthesisRequest{
					Text:    "Sintetizare română",
					VoiceID: "xtts_ro_eduard",
				}

				audioData, err := adapter.Synthesize(req)
				
				Expect(err).ToNot(HaveOccurred())
				Expect(audioData).To(Equal(tts.AudioStream("WAVE_AUDIO_BYTES_MOCK")))
			})
		})
	})
})
