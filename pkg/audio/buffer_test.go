package audio_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/doITmagic/speak4me/pkg/audio"
	"github.com/doITmagic/speak4me/pkg/tts/adapters/mock"
)

var _ = Describe("Audio Streaming", func() {

	Describe("StreamChunker", func() {
		It("trebuie să separe textul corect în propoziții pe baza semnelor de punctuație", func() {
			text := "Salut! Acesta este primul test. Cum funcționează? Perfect."
			
			sentences := audio.SplitTextIntoSentences(text)
			
			Expect(len(sentences)).To(Equal(4))
			Expect(sentences[0]).To(Equal("Salut!"))
			Expect(sentences[1]).To(Equal("Acesta este primul test."))
			Expect(sentences[2]).To(Equal("Cum funcționează?"))
			Expect(sentences[3]).To(Equal("Perfect."))
		})
	})

	Describe("RollingBuffer", func() {
		It("trebuie să mențină ordinea strictă a audio-ului chiar și atunci când TTS-ul returnează asincron", func() {
			mockProvider := mock.NewMockTTS()
			_ = mockProvider.Init(map[string]interface{}{"delay_ms": 10}) // Delay mic pentru test
			
			buffer := audio.NewRollingBuffer(mockProvider, "mock_ro_1")

			inputChan := make(chan string, 10)
			inputChan <- "Propoziția 1 lungă, deci durează."
			inputChan <- "Prop scurtă."
			close(inputChan)

			audioChan, errChan := buffer.ProcessStream(inputChan)

			// Consumăm
			var audioData [][]byte
			for a := range audioChan {
				audioData = append(audioData, a)
			}

			// Nicio eroare pe stream
			Expect(len(errChan)).To(Equal(0))

			// Trebuie să avem 2 chunk-uri audio exact în ordinea input-ului
			Expect(len(audioData)).To(Equal(2))
			Expect(string(audioData[0])).To(ContainSubstring("mock_ro_1"))
			Expect(string(audioData[1])).To(ContainSubstring("mock_ro_1"))
		})
	})
})
