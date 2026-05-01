package tts

// VoiceProfile definește un profil de voce disponibil în sistem
type VoiceProfile struct {
	ID       string   `json:"id"`
	ModelID  string   `json:"model_id"`
	Language string   `json:"language"`
	Name     string   `json:"name"`
	Tags     []string `json:"tags"`
}

// SynthesisRequest definește cerința pentru generarea audio
type SynthesisRequest struct {
	Text     string `json:"text"`
	VoiceID  string `json:"voice_id"`
	Language string `json:"language,omitempty"`
	Stream   bool   `json:"stream"`
}

// AudioStream definește un format generic de răspuns audio
// Momentan un slice de bytes, dar poate fi scalat spre channels pentru streaming
type AudioStream []byte

// TTSProvider este interfața obligatorie pentru orice adaptor ML/Engine
type TTSProvider interface {
	Init(config map[string]interface{}) error
	GetAvailableVoices() ([]VoiceProfile, error)
	Synthesize(req SynthesisRequest) (AudioStream, error)
}
