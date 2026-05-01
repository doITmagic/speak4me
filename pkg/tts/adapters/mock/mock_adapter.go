package mock

import (
	"errors"
	"time"

	"github.com/doITmagic/speak4me/pkg/tts"
)

var ErrVoiceNotFound = errors.New("voice not found in mock")

// MockTTS este un adaptor fals (dummy engine) pentru a simula un TTS real.
// Utilitar pentru testarea arhitecturii sistemului fără modele greoaie în memorie.
type MockTTS struct {
	delayMs int
	voices  []tts.VoiceProfile
}

// NewMockTTS creează un adaptor mock și îi atașează o listă de voci hardcodate
func NewMockTTS() *MockTTS {
	return &MockTTS{
		delayMs: 500, // delay implicit
		voices: []tts.VoiceProfile{
			{
				ID:       "mock_ro_1",
				ModelID:  "mock_v1",
				Language: "ro-RO",
				Name:     "Mock Robot RO",
				Tags:     []string{"mock", "robot", "ro"},
			},
			{
				ID:       "mock_en_1",
				ModelID:  "mock_v1",
				Language: "en-US",
				Name:     "Mock Robot EN",
				Tags:     []string{"mock", "robot", "en"},
			},
		},
	}
}

// Init setează configurația adaptorului (în principal, delay-ul artificial)
func (m *MockTTS) Init(config map[string]interface{}) error {
	if config != nil {
		if delayVal, ok := config["delay_ms"]; ok {
			if delay, isInt := delayVal.(int); isInt {
				m.delayMs = delay
			}
		}
	}
	return nil
}

// GetAvailableVoices returnează vocile suportate de acest mock
func (m *MockTTS) GetAvailableVoices() ([]tts.VoiceProfile, error) {
	return m.voices, nil
}

// Synthesize simulează procesul de Text-to-Speech introducând o întârziere artificială
func (m *MockTTS) Synthesize(req tts.SynthesisRequest) (tts.AudioStream, error) {
	// 1. Verificăm existența vocii
	voiceFound := false
	for _, v := range m.voices {
		if v.ID == req.VoiceID {
			voiceFound = true
			break
		}
	}

	if !voiceFound {
		return nil, ErrVoiceNotFound
	}

	// 2. Simulăm procesarea blocantă
	time.Sleep(time.Duration(m.delayMs) * time.Millisecond)

	// 3. Returnăm un fișier audio fals (aici un byte array binar pentru a trece testul)
	return tts.AudioStream([]byte("mock_audio_data_for_" + req.VoiceID)), nil
}
