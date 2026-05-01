# Task 2: Adaptorul de Referință (Mock TTS Adapter)

## Obiectiv Principal
Crearea unui "dummy engine" care implementează `TTSProvider`. Acesta permite testarea end-to-end a aplicației, a funcțiilor API și a fluxului audio fără a necesita descărcarea sau încărcarea în RAM/VRAM a unui model ML greoi. Orice AI poate folosi acest mock pentru a asigura stabilitatea logică a sistemului.

## Pre-rechizite
- Task 1 finalizat (interfața `TTSProvider` și `ModelRegistry` trebuie să existe în `pkg/tts`).

## Pași de Execuție (Step-by-Step)

### Pasul 2.1: Crearea Structurii Fișierelor
1. Se va crea directorul `pkg/tts/adapters/mock`.
2. Se vor crea fișierele:
   - `mock_adapter.go`
   - `mock_adapter_test.go`

### Pasul 2.2: Implementarea Logică a Mock-ului
1. În `mock_adapter.go` definiți `MockTTS` (struct) care implementează `TTSProvider`.
2. **Init()**: Poate accesa o configurare, de exemplu `delay_ms` pentru a simula timpul de generare a vocii (default 500ms).
3. **GetAvailableVoices()**: Trebuie să returneze o listă hardcodată de 2 voci (ex: ID "mock_ro_1", Name "Mock Robot RO", Language "ro-RO", ModelID "mock_v1").
4. **Synthesize()**: 
   - Trebuie să valideze dacă `req.VoiceID` există în lista lui. Dacă nu, returnează eroare "Voice not found in mock".
   - Dacă există, va introduce un `time.Sleep` (bazat pe `delay_ms`) pentru a simula latența.
   - Returnează un `[]byte` simplu (ex: un fișier wav static încărcat anterior sau un șir de bytes dummy).

### Pasul 2.3: Scrierea Testelor BDD (Ginkgo & Gomega)
1. În directorul `pkg/tts/adapters/mock/` rulați `ginkgo bootstrap` și `ginkgo generate mock_adapter`.
2. **Cazuri de test obligatorii:**
   - [ ] Când `Synthesize` este apelat cu un `VoiceID` corect ("mock_ro_1"), întoarce o eroare nil și lungimea array-ului de bytes este > 0.
   - [ ] Când `Synthesize` este apelat cu un `VoiceID` incorect, întoarce o eroare "Voice not found".
   - [ ] Funcția simulează un delay măsurabil (așteaptă cel puțin `delay_ms`).

## Criterii de Acceptanță (Definition of Done)
- [ ] Interfața `TTSProvider` este perfect satisfăcută de pointer-ul la `MockTTS`.
- [ ] Suita `ginkgo -r pkg/tts/adapters/mock/` raportează 0 erori.
- [ ] Mock-ul a fost înregistrat cu succes într-un `ModelRegistry` în cadrul unui test global de integrare.
