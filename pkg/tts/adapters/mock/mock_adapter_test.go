package mock_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/doITmagic/speak4me/pkg/tts"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/mock"
)

var _ = Describe("MockAdapter", func() {
	var adapter *mock.MockTTS

	BeforeEach(func() {
		adapter = mock.NewMockTTS()
		// Setăm un delay mai mic pentru teste pentru a nu încetini suita de BDD prea tare
		err := adapter.Init(map[string]interface{}{"delay_ms": 100})
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("Inițializare și listare voci", func() {
		It("trebuie să returneze cel puțin vocea 'mock_ro_1'", func() {
			voices, err := adapter.GetAvailableVoices()
			Expect(err).ToNot(HaveOccurred())
			Expect(len(voices)).To(BeNumerically(">", 0))

			found := false
			for _, v := range voices {
				if v.ID == "mock_ro_1" {
					found = true
					break
				}
			}
			Expect(found).To(BeTrue(), "Trebuie să expună vocea de test mock_ro_1")
		})
	})

	Describe("Synthesize", func() {
		Context("când este apelat cu un VoiceID corect", func() {
			It("trebuie să returneze date audio valide", func() {
				req := tts.SynthesisRequest{
					Text:    "Testare mock audio",
					VoiceID: "mock_ro_1",
				}

				start := time.Now()
				audioBytes, err := adapter.Synthesize(req)
				elapsed := time.Since(start)

				Expect(err).ToNot(HaveOccurred())
				Expect(len(audioBytes)).To(BeNumerically(">", 0))
				
				// Verificăm dacă s-a aplicat întârzierea artificială (100ms configurat)
				// Lăsăm o marjă de eroare de +/- 10ms din cauza schedulerului SO
				Expect(elapsed).To(BeNumerically(">=", 90*time.Millisecond))
			})
		})

		Context("când este apelat cu un VoiceID inexistent", func() {
			It("trebuie să returneze o eroare specifică (ErrVoiceNotFound)", func() {
				req := tts.SynthesisRequest{
					Text:    "Acest text nu va fi sintetizat niciodată",
					VoiceID: "voce_invalida_999",
				}

				audioBytes, err := adapter.Synthesize(req)

				Expect(err).To(MatchError(mock.ErrVoiceNotFound))
				Expect(audioBytes).To(BeNil())
			})
		})
	})
})
