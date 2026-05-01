package tts_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/doITmagic/speak4me/pkg/tts"
)

// mockTestProvider este o implementare dummy doar pentru testarea registry-ului
type mockTestProvider struct {
	voices []tts.VoiceProfile
}

func (m *mockTestProvider) Init(config map[string]interface{}) error {
	return nil
}

func (m *mockTestProvider) GetAvailableVoices() ([]tts.VoiceProfile, error) {
	return m.voices, nil
}

func (m *mockTestProvider) Synthesize(req tts.SynthesisRequest) (tts.AudioStream, error) {
	return tts.AudioStream("mock audio data"), nil
}

var _ = Describe("ModelRegistry", func() {
	var registry *tts.ModelRegistry

	BeforeEach(func() {
		// Se rulează înaintea fiecărui test din acest bloc
		registry = tts.NewRegistry()
	})

	Context("Când se înregistrează modele (Providers)", func() {
		It("trebuie să returneze modelul corect fără eroare", func() {
			mockProvider := &mockTestProvider{}
			err := registry.RegisterProvider("model_1", mockProvider)
			Expect(err).ToNot(HaveOccurred())

			retrievedProvider, err := registry.GetProvider("model_1")
			Expect(err).ToNot(HaveOccurred())
			Expect(retrievedProvider).To(Equal(mockProvider))
		})

		It("trebuie să întoarcă ErrModelNotFound pentru un model inexistent", func() {
			retrievedProvider, err := registry.GetProvider("model_inexistent")
			Expect(err).To(MatchError(tts.ErrModelNotFound))
			Expect(retrievedProvider).To(BeNil())
		})
	})

	Context("Când se cere lista tuturor vocilor (ListAllVoices)", func() {
		It("trebuie să unească vocile tuturor provider-elor înregistrate", func() {
			provider1 := &mockTestProvider{
				voices: []tts.VoiceProfile{
					{ID: "v1", Name: "Voice 1", ModelID: "p1"},
					{ID: "v2", Name: "Voice 2", ModelID: "p1"},
				},
			}
			provider2 := &mockTestProvider{
				voices: []tts.VoiceProfile{
					{ID: "v3", Name: "Voice 3", ModelID: "p2"},
				},
			}

			_ = registry.RegisterProvider("p1", provider1)
			_ = registry.RegisterProvider("p2", provider2)

			allVoices, err := registry.ListAllVoices()
			Expect(err).ToNot(HaveOccurred())
			
			// Ne așeteptăm la suma vocilor (2 + 1 = 3)
			Expect(allVoices).To(HaveLen(3))

			// Verificăm sumar existența ID-urilor
			var voiceIDs []string
			for _, v := range allVoices {
				voiceIDs = append(voiceIDs, v.ID)
			}
			Expect(voiceIDs).To(ConsistOf("v1", "v2", "v3"))
		})
	})
})
