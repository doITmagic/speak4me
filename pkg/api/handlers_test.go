package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/doITmagic/speak4me/pkg/api"
	"github.com/doITmagic/speak4me/pkg/db"
	"github.com/doITmagic/speak4me/pkg/tts"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/mock"
)

var _ = Describe("API Handlers", func() {
	var server *api.APIServer
	var registry *tts.ModelRegistry

	BeforeEach(func() {
		// Init registry and mock
		registry = tts.NewRegistry()
		mockAdapter := mock.NewMockTTS()
		_ = mockAdapter.Init(map[string]interface{}{"delay_ms": 10}) // delay foarte mic
		_ = registry.RegisterProvider("mock_v1", mockAdapter)

		// Init In-Memory SQLite Database
		database, err := db.InitDB(":memory:")
		Expect(err).ToNot(HaveOccurred())

		// Initialize Server
		server = api.NewServer(registry, database)
	})

	Describe("GET /api/v1/voices", func() {
		It("trebuie să returneze lista de voci în format JSON", func() {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/voices", nil)
			rr := httptest.NewRecorder()

			server.Router.ServeHTTP(rr, req)

			Expect(rr.Code).To(Equal(http.StatusOK))
			Expect(rr.Header().Get("Content-Type")).To(Equal("application/json"))

			var voices []tts.VoiceProfile
			err := json.NewDecoder(rr.Body).Decode(&voices)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(voices)).To(BeNumerically(">", 0))
			Expect(voices[0].ID).To(Equal("mock_ro_1"))
		})
	})

	Describe("POST /api/v1/synthesize", func() {
		Context("Când request-ul este valid", func() {
			It("trebuie să returneze 200 OK și date audio", func() {
				payload := tts.SynthesisRequest{
					Text:    "Test API",
					VoiceID: "mock_ro_1",
				}
				bodyBytes, _ := json.Marshal(payload)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/synthesize", bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()

				server.Router.ServeHTTP(rr, req)

				Expect(rr.Code).To(Equal(http.StatusOK))
				Expect(rr.Header().Get("Content-Type")).To(Equal("audio/wav"))
				Expect(rr.Body.Len()).To(BeNumerically(">", 0))
			})
		})

		Context("Când lipsește VoiceID", func() {
			It("trebuie să returneze 400 Bad Request", func() {
				payload := tts.SynthesisRequest{
					Text: "Test API fara voce",
				}
				bodyBytes, _ := json.Marshal(payload)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/synthesize", bytes.NewBuffer(bodyBytes))
				rr := httptest.NewRecorder()

				server.Router.ServeHTTP(rr, req)

				Expect(rr.Code).To(Equal(http.StatusBadRequest))
			})
		})

		Context("Când vocea nu există", func() {
			It("trebuie să returneze 404 Not Found", func() {
				payload := tts.SynthesisRequest{
					Text:    "Test API voce invalida",
					VoiceID: "voce_fantoma",
				}
				bodyBytes, _ := json.Marshal(payload)

				req := httptest.NewRequest(http.MethodPost, "/api/v1/synthesize", bytes.NewBuffer(bodyBytes))
				rr := httptest.NewRecorder()

				server.Router.ServeHTTP(rr, req)

				Expect(rr.Code).To(Equal(http.StatusNotFound))
			})
		})
	})
})
