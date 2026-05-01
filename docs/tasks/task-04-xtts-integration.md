# Task 4: Integrare Model Principal XTTS-v2-Romanian

## Obiectiv Principal
Acest task implică complexitatea centrală: integrarea unui motor AI masiv (`eduardem/xtts-v2-romanian`) în aplicația noastră de tip "Registry". Modelul ML necesită atenție sporită pentru managementul memoriei (RAM/VRAM) și performanța de inferență.

## Pre-rechizite
- Un mediu de test unde XTTS poate fi compilat/executat (având în vedere că modelele Coqui/XTTS cer deseori backend C++ (ONNX) sau Python).
- **Decizie Arhitecturală Asumată:** Deoarece aplicația e scrisă în Go, integrarea cu XTTS se va face probabil fie prin rularea unui mic sidecar process (ex: binar Python compilat sau server Python local), fie prin linkare directă cu o librărie C/C++ ONNX Runtime sau un port Rust (ținând cont de experiența *speak2me*). **AI-ul care implementează acest task va trebui să verifice structura existentă sau să scrie FFI (Foreign Function Interface) către motorul XTTS ales.**

## Pași de Execuție (Step-by-Step)

### Pasul 4.1: Adaptorul XTTS (`pkg/tts/adapters/xtts/`)
1. Creați `xtts_adapter.go`.
2. Structura `XTTSAdapter` care implementează `TTSProvider`.
3. **Init()**: Trebuie să valideze prezența fișierelor modelului (ex: `model.bin`, `config.json`, `vocab.json` pentru `eduardem/xtts-v2-romanian`) pe disc. Dacă nu există, poate declanșa un downloader (funcționalitate adăugată suplimentar).
4. Setează conexiunea către procesul de inferență sau instanțiază contextul ONNX/Rust/CGO.

### Pasul 4.2: Mapping-ul Vocilor (Speaker Conditioning)
1. XTTS v2 este capabil de voice cloning (speaker conditioning). Adaptorul trebuie să citească dintr-un folder un set de wav-uri de referință (ex: `voices/maria.wav`, `voices/ion.wav`).
2. Funcția `GetAvailableVoices()` va citi acest folder și va expune `VoiceProfile` pentru fiecare sample găsit, indicând limba `ro-RO`.

### Pasul 4.3: Implementarea Synthesize()
1. Funcția primește textul, identifică vocea (fișierul de referință) din memorie.
2. Formatează inputul pentru motorul de inferență XTTS.
3. Preia array-ul de floating-point (audio raw) sau bytes (wav) generat de motor și îl convertește în formatul unificat al interfeței (ex: standardizat la PCM 16-bit 24kHz Mono).

### Pasul 4.4: Testare BDD pentru XTTS
1. Suita Ginkgo de aici are un caracter special. Testele XTTS pot fi greoaie și pot pica dacă PC-ul nu are resurse.
2. **Implementați Tag-uri Ginkgo:** Adăugați tag-ul `//go:build integration` sau rulați doar dacă o variabilă de environment `RUN_XTTS_TESTS=1` este activă.
3. **Scenariu:** Rulați o sintetizare scurtă de 2-3 cuvinte. Verificați că funcția întoarce array-ul binar, fără a depăși o latență de X secunde (timeout strict).

## Criterii de Acceptanță (Definition of Done)
- [ ] Adaptorul încarcă valid un model HuggingFace compatibil local.
- [ ] Synthesize generează audio PCM curat.
- [ ] Procesul nu face memory leak între apeluri de sintetizare.
