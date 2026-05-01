# Specificație de Dezvoltare: ChatTTS_Speaker

Acest document folosește metodologia Spec-Driven Development pentru a defini arhitectura și pașii de implementare pentru a doua aplicație din suita de asistență vocală: **ChatTTS_Speaker**.

## Faza 1: Colectarea Cerințelor (Requirements)

**Context:** ChatTTS_Speaker este o aplicație de tip Text-to-Speech (TTS) capabilă să preia text într-o anumită limbă și să genereze conținut audio vorbit. Aplicația trebuie să fie agnostică din punct de vedere al modelului (să suporte multiple engine-uri și modele), pregătind terenul pentru aplicația #3 care va uni STT (speak2me), un LLM și acest TTS.

### User Stories

1. **Ca utilizator**, vreau să pot trimite un text către sistem și să primesc o redare audio naturală, pentru a putea asculta informația.
2. **Ca utilizator**, vreau să selectez limba și o voce specifică (ex: voce feminină, masculină, ton diferit), pentru a personaliza interacțiunea.
3. **Ca dezvoltator**, vreau o arhitectură de tip "plug-and-play" pentru modele, astfel încât să pot adăuga modele TTS noi (locale, specializate pe limbi, sau chiar API-uri externe) fără să modific logica de bază a aplicației.
4. **Ca sistem integrator (Aplicația #3)**, vreau să pot trimite text pe măsură ce este generat de un LLM (streaming/chuncking) și să primesc un flux audio continuu, pentru a avea o conversație fără latență.

### Criterii de Acceptare (Format EARS)

1. **WHEN** sistemul primește o cerere validă cu text, ID limbă și ID voce **THEN** sistemul **SHALL** selecta modelul adecvat și va genera fluxul audio.
2. **IF** vocea sau modelul solicitat nu sunt disponibile/instalate **THEN** sistemul **SHALL** returna o eroare clară și va face fallback la o voce implicită (dacă este configurată).
3. **WHEN** dezvoltatorul adaugă un nou adaptor de model conform interfeței sistemului **THEN** sistemul **SHALL** expune automat noile voci fără modificări în modulele de rutare.
4. **WHEN** se solicită listarea vocilor **THEN** sistemul **SHALL** returna o listă completă de `VoiceProfiles` agregată de la toate modelele active, incluzând suportul de limbă și capabilitățile.

---

## Faza 2: Documentația de Design (Design Documentation)

### Overview
Sistemul va fi construit în jurul unui **Registry de Modele**, funcționând ca un engine agnostico-model de Text-to-Speech. Logica centrală comunică cu motoarele specifice (ChatTTS, Piper etc.) printr-o interfață standardizată.

### Tech Stack (Aliniat cu `speak2me`)
Pentru a menține uniformitatea arhitecturală cu prima aplicație din serie, se vor folosi următoarele tehnologii:
- **Backend / Core Engine:** **Golang** (pentru gestionarea concurenței, rutarea modelelor și HTTP/gRPC server).
- **System Level (opțional):** **Rust** (dacă este necesar pentru legături cu sisteme low-level de audio sau optimizări de performanță pentru inferență).
- **Frontend / Desktop GUI:** **Wails (React 18 + TypeScript + Vite)** pentru interfața cu utilizatorul, oferind un UI fluid și cross-platform similar cu cel din *speak2me*.
- **Bază de date:** **SQLite** (pentru stocarea profilurilor vocale, setărilor utilizatorilor și configurărilor sistemului, fiind simplu, eficient și suficient pentru moment).
- **Testare (BDD):** **Ginkgo & Gomega** (pentru testare Behaviour-Driven Development a logicii core și a integrării modulelor).

### Arhitectura Sistemului

1. **Frontend Layer (Wails + React):** Interfața utilizatorului pentru selecția vocilor, testare, managementul modelelor și setări.
   - **Flux UI impus:** Selecția se va face strict în cascadă: utilizatorul alege **Limba** (filtrează modelele) -> alege **Modelul** (filtrează vocile din modelul respectiv) -> alege **Vocea**. Acest lucru simplifică UX-ul când numărul de modele și voci crește.
2. **API / Input Layer (Go):** Punctul de intrare (HTTP, gRPC, sau Wails Bindings) care primește cererile.
3. **TTS Manager (Orchestrator - Go):** Inima aplicației. Primește cererea, se uită la `VoiceID` sau `Language`, determină ce model este cel mai potrivit și rutează cererea.
4. **Model Adapters (Interfețe - Go/Rust):** Fiecare motor TTS are un adaptor propriu care implementează o interfață comună (ex: `GenerateAudio(text)`). Adaptorul poate face FFI către Rust sau executa procese.
5. **Audio Output / Streaming Layer:** Gestionează trimiterea binarului audio înapoi către client sau direct către device-ul audio.

### Modele de Date (Data Models)

```go
// Definirea unei voci disponibile în sistem
type VoiceProfile struct {
    ID          string   // Ex: "chattts_ro_female_1"
    ModelID     string   // Ex: "chattts_v1", "piper_local"
    Language    string   // Ex: "ro-RO", "en-US"
    Name        string   // Ex: "Maria"
    Tags        []string // Ex: ["calm", "fast", "news"]
}

// Cererea de sinteză vocală
type SynthesisRequest struct {
    Text        string
    VoiceID     string
    Language    string   // Opțional, util dacă se cere auto-selectare a vocii
    Stream      bool     // Dacă true, returnează chunk-uri audio pe măsură ce sunt gata
}

// Interfața obligatorie pentru orice model TTS nou adăugat
type TTSProvider interface {
    Init(config map[string]interface{}) error
    GetAvailableVoices() ([]VoiceProfile, error)
    Synthesize(req SynthesisRequest) (AudioStream, error)
}
```

### Gestionarea Erorilor (Error Handling)
- **Model Initialization Failure:** Dacă un model nu se poate încărca (ex. lipsă VRAM, fișiere model lipsă), se va loga eroarea, dar sistemul general va porni deservind cereri folosind restul modelelor disponibile.
- **Synthesize Timeout:** Timeout-uri stricte per cerere pentru a preveni blocarea pipeline-ului (vital pentru interacțiunea viitoare cu LLM).

---

## Faza 3: Planificarea Sarcinilor (Task Planning)

Abordare: **Foundation-First** (Mai întâi construim interfețele și managerul, apoi integrăm primul model real).

- [ ] **1. Arhitectura de Bază (Core System)**
  - [ ] 1.1 Definirea structurilor de date (`VoiceProfile`, `SynthesisRequest`) și a interfeței `TTSProvider`.
  - [ ] 1.2 Implementarea `ModelRegistry` (o structură care memorează ce modele sunt înregistrate și mapează vocile la ele).
  - [ ] 1.3 Implementarea funcției de rutare (`Manager.Synthesize()`) care preia un text, găsește modelul potrivit pentru `VoiceID`-ul cerut și apelează adaptorul.

- [ ] **2. Adaptorul de Referință (Dummy/Mock Model)**
  - [ ] 2.1 Crearea unui adaptor Mock care doar generează un fișier audio preînregistrat (sau beep-uri). Acest lucru ne permite să testăm end-to-end API-ul înainte să conectăm un AI greoi.

- [ ] **3. Input / API Layer**
  - [ ] 3.1 Crearea unui endpoint HTTP (ex: `/api/tts/voices`) pentru a lista toate vocile disponibile.
  - [ ] 3.2 Crearea unui endpoint HTTP (ex: `/api/tts/synthesize`) care acceptă JSON și returnează un stream de date `.wav` sau `.mp3`.

- [ ] **4. Integrarea Primului Model TTS Real (XTTS-v2-Romanian)**
  - [ ] 4.1 Dezvoltarea adaptorului specific pentru motorul XTTS (`eduardem/xtts-v2-romanian` de pe HuggingFace).
  - [ ] 4.2 Încărcarea modelelor/parametrilor și expunerea vocilor reale către `ModelRegistry`.
  - [ ] 4.3 Implementarea testelor BDD (Ginkgo) pentru a asigura un flow corect de sintetizare folosind XTTS.

- [ ] **5. Sistemul de Streaming Audio (Pentru Aplicația #3)**
  - [ ] 5.1 Implementarea suportului pentru "chunking" (să proceseze propoziție cu propoziție și să livreze audio continuu), pregătindu-ne pentru input-ul pe bucăți de la un LLM.
