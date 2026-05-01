# Task 1: Arhitectura de Bază și Setarea Proiectului (Core System)

## Obiectiv Principal
Inițializarea proiectului Golang `speak4me`, configurarea suitei de testare BDD (Ginkgo & Gomega) și definirea tipurilor de date fundamentale și a `ModelRegistry`-ului. Orice agent AI va folosi acest fișier pentru a pune bazele codului independent de platformă.

## Pre-rechizite
- Golang (recomandat v1.21+) instalat.
- Cunoștințe de concurență în Go (`sync.RWMutex`).

## Pași de Execuție (Step-by-Step)

### Pasul 1.1: Inițializarea Proiectului și a Dependințelor
1. Se rulează `go mod init github.com/doITmagic/speak4me` (dacă nu există deja `go.mod`).
2. Se instalează framework-ul de testare:
   ```bash
   go get github.com/onsi/ginkgo/v2/ginkgo
   go get github.com/onsi/gomega/...
   ```

### Pasul 1.2: Definirea Tipurilor (Structurilor) de Bază
1. Se va crea fișierul `pkg/tts/types.go`.
2. Trebuie să conțină definirea exactă a următoarelor elemente:
   - `VoiceProfile` (struct): `ID` (string), `ModelID` (string), `Language` (string), `Name` (string), `Tags` ([]string).
   - `SynthesisRequest` (struct): `Text` (string), `VoiceID` (string), `Language` (string), `Stream` (bool).
   - `AudioStream` (type): Momentan un `[]byte` sau un channel `<-chan []byte` pentru viitorul streaming.
   - `TTSProvider` (interface):
     - `Init(config map[string]interface{}) error`
     - `GetAvailableVoices() ([]VoiceProfile, error)`
     - `Synthesize(req SynthesisRequest) ([]byte, error)`

### Pasul 1.3: Implementarea Model Registry
1. Se va crea fișierul `pkg/tts/registry.go`.
2. Se va defini structura `ModelRegistry` care conține un `map[string]TTSProvider` (sau similar) și un `sync.RWMutex` pentru operații thread-safe.
3. Metode publice obligatorii:
   - `NewRegistry() *ModelRegistry`
   - `RegisterProvider(modelID string, provider TTSProvider) error`
   - `GetProvider(modelID string) (TTSProvider, error)`
   - `ListAllVoices() ([]VoiceProfile, error)` (iterează prin toate provider-ele și adună vocile).

### Pasul 1.4: Scriere Teste BDD
1. Se generează suita de teste: `cd pkg/tts && ginkgo bootstrap && ginkgo generate registry`.
2. Se editează `registry_test.go` folosind `Describe`, `Context`, `It`.
3. **Cazuri de test obligatorii:**
   - [x] Când se înregistrează un provider valid, funcția `GetProvider` trebuie să îl returneze fără eroare.
   - [x] Când se cere un model inexistent, `GetProvider` trebuie să întoarcă o eroare clară de tip "ModelNotFound".
   - [x] Când 2 provideri raportează voci, `ListAllVoices` trebuie să le unească și să returneze suma exactă a profilurilor.

## Criterii de Acceptanță (Definition of Done)
- [x] Tot codul se află în `pkg/tts/`.
- [x] Comanda `ginkgo -r pkg/tts` rulează cu succes (0 failures).
- [x] Structurile permit adăugarea ulterioară de modele (sunt complet decuplate de implementarea internă a unui motor AI specific).
