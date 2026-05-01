package audio

import (
	"fmt"
	"github.com/doITmagic/speak4me/pkg/tts"
)

// audioTask reține referința viitoare la rezultatul audio pentru a menține ordinea
type audioTask struct {
	audioChan chan []byte
	errChan   chan error
}

// RollingBuffer preia propoziții, le trimite asincron către TTS și livrează audio în ordine strictă
type RollingBuffer struct {
	provider tts.TTSProvider
	voiceID  string
}

func NewRollingBuffer(provider tts.TTSProvider, voiceID string) *RollingBuffer {
	return &RollingBuffer{
		provider: provider,
		voiceID:  voiceID,
	}
}

// ProcessStream preia canalul de propoziții și returnează un canal unde picură fișierele audio în ordine corectă.
func (rb *RollingBuffer) ProcessStream(sentences <-chan string) (<-chan []byte, <-chan error) {
	outAudio := make(chan []byte, 50)
	outErr := make(chan error, 1)
	
	// Un canal unde punem task-urile în ordinea exactă în care apar propozițiile
	taskQueue := make(chan *audioTask, 100)

	// Goroutine 1: Preia propoziții, le lansează asincron spre generare și le înregistrează în taskQueue
	go func() {
		defer close(taskQueue)
		for sentence := range sentences {
			if sentence == "" {
				continue
			}

			task := &audioTask{
				audioChan: make(chan []byte, 1),
				errChan:   make(chan error, 1),
			}
			taskQueue <- task

			// Worker asincron care generează audio pentru această propoziție specifică
			go func(text string, t *audioTask) {
				audio, err := rb.provider.Synthesize(tts.SynthesisRequest{
					Text:    text,
					VoiceID: rb.voiceID,
				})
				if err != nil {
					t.errChan <- fmt.Errorf("eroare sintetizare '%s': %w", text, err)
					close(t.audioChan)
				} else {
					t.audioChan <- audio
					close(t.audioChan)
				}
				close(t.errChan)
			}(sentence, task)
		}
	}()

	// Goroutine 2: Consumă taskQueue strict în ordine și face forward către stream-ul final HTTP
	go func() {
		defer close(outAudio)
		defer close(outErr)

		for task := range taskQueue {
			// Citim audio-ul — dacă workerul a avut eroare, audioChan va fi închis fără date
			audio, ok := <-task.audioChan
			if ok && audio != nil {
				outAudio <- audio
			} else {
				// Dacă nu am primit audio, verificăm eroarea
				if err, hasErr := <-task.errChan; hasErr && err != nil {
					outErr <- err
					return
				}
			}
		}
	}()

	return outAudio, outErr
}
