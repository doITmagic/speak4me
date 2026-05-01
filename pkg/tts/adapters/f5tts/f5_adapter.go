package f5tts

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/doITmagic/speak4me/pkg/tts"
	ort "github.com/yalue/onnxruntime_go"
)

const (
	ModelID    = "cdorob/f5-tts-romanian"
	SampleRate = 24000

	HOP_LENGTH     = 256
	NUM_HEAD       = 16
	HEAD_DIM       = 64
	N_MELS         = 100
	TEXT_EMBED_LEN = 612 // 512 + 100
	NFE_STEP       = 32
	FUSE_NFE       = 1
	SPEED          = 1.0
)

type F5TTSAdapter struct {
	modelsDir string
	voicesDir string
	vocab     map[string]int32

	preprocess  *ort.DynamicAdvancedSession
	transformer *ort.DynamicAdvancedSession
	decode      *ort.DynamicAdvancedSession

	mu        sync.Mutex
	isLoaded  bool
	ortInited bool
}

func NewF5TTSAdapter() *F5TTSAdapter {
	return &F5TTSAdapter{
		modelsDir: "./models/f5-tts-romanian",
		voicesDir: "./voices/f5",
		vocab:     make(map[string]int32),
	}
}

func (a *F5TTSAdapter) loadVocab(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(b), "\n")
	for i, l := range lines {
		l = strings.TrimSpace(l) // in vocab, each line is a token
		// However, vocab might have empty string or spaces mapped. 
		// For simplicity, we split raw bytes if needed, but strings.Split is fine.
		if len(l) > 0 {
			a.vocab[l] = int32(i)
		} else if len(lines[i]) > 0 {
			// e.g. a space char
			a.vocab[lines[i]] = int32(i)
		}
	}
	return nil
}

func (a *F5TTSAdapter) Init(config map[string]interface{}) error {
	if config != nil {
		if dir, ok := config["models_dir"].(string); ok && dir != "" {
			a.modelsDir = dir
		}
		if dir, ok := config["voices_dir"].(string); ok && dir != "" {
			a.voicesDir = dir
		}
	}

	_ = os.MkdirAll(a.voicesDir, os.ModePerm)

	// Incarca vocabular
	vocabPath := filepath.Join(a.modelsDir, "vocab.txt")
	if err := a.loadVocab(vocabPath); err != nil {
		return fmt.Errorf("eroare vocab.txt: %w", err)
	}

	requiredFiles := []string{"F5_Preprocess.onnx", "F5_Transformer.onnx", "F5_Decode.onnx"}
	for _, f := range requiredFiles {
		path := filepath.Join(a.modelsDir, f)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("fișierul ONNX lipsește: %s", path)
		}
	}

	if !a.ortInited {
		ortLibPath := findOrtLib()
		if ortLibPath == "" {
			return fmt.Errorf("nu s-a găsit libonnxruntime.so")
		}
		ort.SetSharedLibraryPath(ortLibPath)
		if err := ort.InitializeEnvironment(); err != nil {
			return fmt.Errorf("eroare ORT env: %w", err)
		}
		a.ortInited = true
	}

	opts, e := ort.NewSessionOptions()
	if e != nil {
		return e
	}
	defer opts.Destroy()
	_ = opts.SetIntraOpNumThreads(0)
	_ = opts.SetInterOpNumThreads(0)

	var err error
	a.preprocess, err = ort.NewDynamicAdvancedSession(
		filepath.Join(a.modelsDir, "F5_Preprocess.onnx"),
		[]string{"audio", "text_ids", "max_duration"},
		[]string{"noise", "rope_cos_q", "rope_sin_q", "rope_cos_k", "rope_sin_k", "cat_mel_text", "cat_mel_text_drop", "ref_signal_len"},
		opts)
	if err != nil {
		return fmt.Errorf("load Preprocess failed: %w", err)
	}

	a.transformer, err = ort.NewDynamicAdvancedSession(
		filepath.Join(a.modelsDir, "F5_Transformer.onnx"),
		[]string{"noise", "rope_cos_q", "rope_sin_q", "rope_cos_k", "rope_sin_k", "cat_mel_text", "cat_mel_text_drop", "time_step.1"},
		[]string{"denoised", "time_step"},
		opts)
	if err != nil {
		return fmt.Errorf("load Transformer failed: %w", err)
	}

	a.decode, err = ort.NewDynamicAdvancedSession(
		filepath.Join(a.modelsDir, "F5_Decode.onnx"),
		[]string{"denoised", "ref_signal_len"},
		[]string{"output_audio"},
		opts)
	if err != nil {
		return fmt.Errorf("load Decode failed: %w", err)
	}

	a.isLoaded = true
	return nil
}

func (a *F5TTSAdapter) GetAvailableVoices() ([]tts.VoiceProfile, error) {
	var voices []tts.VoiceProfile

	files, err := os.ReadDir(a.voicesDir)
	if err != nil {
		return voices, nil
	}

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(strings.ToLower(f.Name()), ".wav") {
			name := strings.TrimSuffix(f.Name(), filepath.Ext(f.Name()))
			voices = append(voices, tts.VoiceProfile{
				ID:       "f5_ro_" + name,
				ModelID:  ModelID,
				Language: "ro-RO",
				Name:     strings.Title(name),
				Tags:     []string{"f5-tts", "romanian", "onnx"},
			})
		}
	}

	if len(voices) == 0 {
		voices = append(voices, tts.VoiceProfile{
			ID:       "f5_ro_default",
			ModelID:  ModelID,
			Language: "ro-RO",
			Name:     "F5-TTS Romanian",
			Tags:     []string{"f5-tts", "romanian", "onnx"},
		})
	}

	return voices, nil
}

// Mapare caractere latine (din python logic)
func (a *F5TTSAdapter) textToIds(text string) []int32 {
	var ids []int32
	clean := text

	for _, char := range clean {
		c := string(char)
		if id, ok := a.vocab[c]; ok {
			ids = append(ids, id)
		} else {
			// space sau unknown = 0 in multe cazuri
			ids = append(ids, 0)
		}
	}
	return ids
}

func countPunc(text string) int {
	count := 0
	puncs := []string{"。", "，", "、", "；", "：", "？", "！", ".", ",", ";", ":", "?", "!"}
	for _, p := range puncs {
		count += strings.Count(text, p)
	}
	return count
}

func (a *F5TTSAdapter) Synthesize(req tts.SynthesisRequest) (tts.AudioStream, error) {
	if !a.isLoaded {
		return nil, fmt.Errorf("F5 nu este initializat")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	fmt.Printf("[F5-TTS] Incep generarea nativa ONNX pentru: %s\n", req.Text)
	startTotal := time.Now()

	// Extragem numele fisierului din VoiceID
	// De exemplu, din "f5_ro_f5_ro_default" -> "f5_ro_default"
	voiceName := strings.TrimPrefix(req.VoiceID, "f5_ro_")
	if voiceName == "" {
		voiceName = "f5_ro_default"
	}

	refPath := filepath.Join(a.voicesDir, voiceName+".wav")
	txtPath := filepath.Join(a.voicesDir, voiceName+".txt")

	fmt.Printf("DEBUG: Loading refPath: %s\n", refPath)
	wavBytes, err := os.ReadFile(refPath)
	if err != nil {
		return nil, fmt.Errorf("ref audio lipsa %s: %w", refPath, err)
	}
	
	refTextBytes, err := os.ReadFile(txtPath)
	if err != nil {
		return nil, fmt.Errorf("ref text lipsa %s: %w", txtPath, err)
	}
	refText := strings.TrimSpace(string(refTextBytes))

	audioSamples := make([]int16, 0)
	if len(wavBytes) > 44 {
		for i := 44; i < len(wavBytes)-1; i += 2 {
			audioSamples = append(audioSamples, int16(binary.LittleEndian.Uint16(wavBytes[i:i+2])))
		}
	}

	genText := req.Text
	fullText := refText + genText

	refTextLen := len([]byte(refText)) + 3*countPunc(refText)
	genTextLen := len([]byte(genText)) + 3*countPunc(genText)
	if refTextLen == 0 {
		refTextLen = 1
	}

	audioLen := int64(len(audioSamples))
	refAudioLen := audioLen/HOP_LENGTH + 1
	maxDuration := refAudioLen + int64(float64(refAudioLen)/float64(refTextLen)*float64(genTextLen)/SPEED)

	textIds := a.textToIds(fullText)
	textIdsLen := int64(len(textIds))

	// Functie helper pt siguranta
	var tensorsToDestroy []func()
	defer func() {
		for _, destroyFn := range tensorsToDestroy {
			destroyFn()
		}
	}()

	addTensor := func(t *ort.Tensor[float32], e error) (*ort.Tensor[float32], error) {
		if t != nil {
			tensorsToDestroy = append(tensorsToDestroy, func() { t.Destroy() })
		}
		return t, e
	}
	addTensorI32 := func(t *ort.Tensor[int32], e error) (*ort.Tensor[int32], error) {
		if t != nil {
			tensorsToDestroy = append(tensorsToDestroy, func() { t.Destroy() })
		}
		return t, e
	}
	addTensorI64 := func(t *ort.Tensor[int64], e error) (*ort.Tensor[int64], error) {
		if t != nil {
			tensorsToDestroy = append(tensorsToDestroy, func() { t.Destroy() })
		}
		return t, e
	}
	addTensorI16 := func(t *ort.Tensor[int16], e error) (*ort.Tensor[int16], error) {
		if t != nil {
			tensorsToDestroy = append(tensorsToDestroy, func() { t.Destroy() })
		}
		return t, e
	}

	tAudio, err := addTensorI16(ort.NewTensor(ort.NewShape(1, 1, audioLen), audioSamples))
	if err != nil { return nil, fmt.Errorf("tAudio err: %v", err) }
	tTextIds, err := addTensorI32(ort.NewTensor(ort.NewShape(1, textIdsLen), textIds))
	if err != nil { return nil, fmt.Errorf("tTextIds err: %v", err) }
	tMaxDur, err := addTensorI64(ort.NewTensor(ort.NewShape(1), []int64{maxDuration}))
	if err != nil { return nil, fmt.Errorf("tMaxDur err: %v", err) }

	tNoise, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(1, maxDuration, N_MELS)))
	if err != nil { return nil, fmt.Errorf("tNoise err: %v", err) }
	tRopeCosQ, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(2, NUM_HEAD, maxDuration, HEAD_DIM)))
	if err != nil { return nil, fmt.Errorf("tRopeCosQ err: %v", err) }
	tRopeSinQ, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(2, NUM_HEAD, maxDuration, HEAD_DIM)))
	if err != nil { return nil, fmt.Errorf("tRopeSinQ err: %v", err) }
	tRopeCosK, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(2, NUM_HEAD, HEAD_DIM, maxDuration)))
	if err != nil { return nil, fmt.Errorf("tRopeCosK err: %v", err) }
	tRopeSinK, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(2, NUM_HEAD, HEAD_DIM, maxDuration)))
	if err != nil { return nil, fmt.Errorf("tRopeSinK err: %v", err) }
	tCatMelText, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(1, maxDuration, TEXT_EMBED_LEN)))
	if err != nil { return nil, fmt.Errorf("tCatMelText err: %v", err) }
	tCatMelDrop, err := addTensor(ort.NewEmptyTensor[float32](ort.NewShape(1, maxDuration, TEXT_EMBED_LEN)))
	if err != nil { return nil, fmt.Errorf("tCatMelDrop err: %v", err) }
	tRefSigLen, err := addTensorI64(ort.NewEmptyTensor[int64](ort.NewShape()))
	if err != nil { return nil, fmt.Errorf("tRefSigLen err: %v", err) }

	inputsA := []ort.Value{tAudio, tTextIds, tMaxDur}
	outputsA := []ort.Value{tNoise, tRopeCosQ, tRopeSinQ, tRopeCosK, tRopeSinK, tCatMelText, tCatMelDrop, tRefSigLen}

	fmt.Println("[F5-TTS] Rulăm Preprocess...")
	if err := a.preprocess.Run(inputsA, outputsA); err != nil {
		return nil, fmt.Errorf("preprocess error: %w", err)
	}

	tTimeStep, err := addTensorI32(ort.NewTensor(ort.NewShape(1), []int32{0}))
	if err != nil { return nil, fmt.Errorf("tTimeStep err: %v", err) }
	
	fmt.Printf("[F5-TTS] Rulăm Transformer (32 NFE steps)...\n")
	for i := 0; i < NFE_STEP-2; i += FUSE_NFE {
		tDenoised, err := ort.NewEmptyTensor[float32](ort.NewShape(1, maxDuration, N_MELS))
		if err != nil { return nil, fmt.Errorf("tDenoised err: %v", err) }

		inputsB := []ort.Value{tNoise, tRopeCosQ, tRopeSinQ, tRopeCosK, tRopeSinK, tCatMelText, tCatMelDrop, tTimeStep}
		outputsB := []ort.Value{tDenoised, tTimeStep}
		
		fmt.Printf("Step %d, tTimeStep: %v\n", i, tTimeStep.GetData())

		if err := a.transformer.Run(inputsB, outputsB); err != nil {
			tDenoised.Destroy()
			return nil, fmt.Errorf("transformer error at step %d: %w", i, err)
		}

		copy(tNoise.GetData(), tDenoised.GetData())
		tDenoised.Destroy()
	}

	refSignalLenVal := tRefSigLen.GetData()[0]
	generatedLen := (maxDuration - refSignalLenVal - 1) * HOP_LENGTH

	fmt.Printf("DEBUG Decode: maxDuration=%d, refSignalLenVal=%d, generatedLen=%d\n", maxDuration, refSignalLenVal, generatedLen)

	tGenAudio, err := addTensorI16(ort.NewEmptyTensor[int16](ort.NewShape(1, 1, generatedLen)))
	if err != nil { return nil, fmt.Errorf("tGenAudio err: %v", err) }

	inputsC := []ort.Value{tNoise, tRefSigLen}
	outputsC := []ort.Value{tGenAudio}

	fmt.Println("[F5-TTS] Rulăm Vocoder Decode...")
	if err := a.decode.Run(inputsC, outputsC); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	rawAudio := tGenAudio.GetData()

	fmt.Printf("[F5-TTS] Generare completă în %v. Rezultat: %d samples.\n", time.Since(startTotal), len(rawAudio))
	return createWavBytes(rawAudio), nil
}

func createWavBytes(samples []int16) []byte {
	numSamples := len(samples)
	dataSize := numSamples * 2
	wav := make([]byte, 44+dataSize)

	copy(wav[0:4], "RIFF")
	binary.LittleEndian.PutUint32(wav[4:8], uint32(36+dataSize))
	copy(wav[8:12], "WAVE")
	copy(wav[12:16], "fmt ")
	binary.LittleEndian.PutUint32(wav[16:20], 16)
	binary.LittleEndian.PutUint16(wav[20:22], 1)
	binary.LittleEndian.PutUint16(wav[22:24], 1)
	binary.LittleEndian.PutUint32(wav[24:28], uint32(SampleRate))
	binary.LittleEndian.PutUint32(wav[28:32], uint32(SampleRate*2))
	binary.LittleEndian.PutUint16(wav[32:34], 2)
	binary.LittleEndian.PutUint16(wav[34:36], 16)
	copy(wav[36:40], "data")
	binary.LittleEndian.PutUint32(wav[40:44], uint32(dataSize))

	for i := 0; i < numSamples; i++ {
		offset := 44 + i*2
		binary.LittleEndian.PutUint16(wav[offset:offset+2], uint16(samples[i]))
	}

	return wav
}

func findOrtLib() string {
	if p := os.Getenv("ORT_LIB_PATH"); p != "" {
		return p
	}
	var candidates []string
	switch runtime.GOOS {
	case "linux":
		candidates = []string{
			"/usr/lib/libonnxruntime.so",
			"/usr/local/lib/libonnxruntime.so",
			"/usr/lib/x86_64-linux-gnu/libonnxruntime.so",
			"./libonnxruntime.so",
		}
	}
	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			return c
		}
	}
	return ""
}
