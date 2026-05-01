# Task 5: Chunking & Audio Streaming (Pregătire LLM)

## Obiectiv Principal
Pentru a atinge o viteză de reacție extrem de rapidă în viitoarea aplicație #3 (unde un LLM generează text pe măsură ce "gândește"), motorul nostru TTS trebuie să proceseze textul pe bucăți (chunk-uri). Odată generată o bucată de audio, aceasta este trimisă înapoi printr-un flux continuu (Rolling Buffer / WebSocket stream), astfel încât vocea să înceapă imediat ce LLM-ul termină prima propoziție, mascând latența totală.

## Pre-rechizite
- Înțelegerea conceptelor de "Goroutine", "Channels", "Server-Sent Events (SSE)" sau "WebSockets".
- Implementarea finalizată a Task-urilor 1, 3 și 4 (un motor capabil de sintetizare).

## Pași de Execuție (Step-by-Step)

### Pasul 5.1: Chunker de Text (`pkg/audio/chunker.go`)
1. Implementați o funcție inteligentă de Split care să nu taie textul arbitrar, ci natural, la nivel de punctuație (. ! ? \n) sau virgulă, luând în calcul limite de lungime (ex: max 20 de cuvinte per bucată pentru a evita OOM în VRAM la XTTS).
2. Trebuie să accepte un stream de cuvinte (ex: `<-chan string`) dinspre o sursă viitoare.

### Pasul 5.2: The Rolling Buffer (`pkg/audio/buffer.go`)
1. Creați un buffer avansat folosind canale în Go: `audioQueue := make(chan []byte, 100)`.
2. Dezvoltați un Worker Pool. Imediat cum un Chunk de text este formatat, se trimite la `XTTSAdapter.Synthesize()`.
3. Rezultatele (PCM/Wav bytes) sunt introduse în canal în ordinea strictă a propozițiilor primite (necesită sincronizare ex. map+index sau Heap).
4. Pe măsură ce bufferele sunt decodate, ele se trimit către Clientul HTTP via chunked transfer (ex: `http.Flusher` sau peste Websocket).

### Pasul 5.3: Adaptarea API-ului (`pkg/api/`)
1. Adăugați o metodă nouă `POST /api/v1/stream_synthesize`.
2. Va menține conexiunea deschisă și va spiona `audioQueue`. De fiecare dată când apar bytes noi de audio valid, face `w.Write(bytes)` urmat de `w.(http.Flusher).Flush()`.

### Pasul 5.4: Testare Funcțională Extremă (BDD Ginkgo)
1. Construiți suita de testare `buffer_test.go` cu Ginkgo.
2. **Cazuri de test obligatorii:**
   - [x] Oferiți un text mare de 10 propoziții. Verificați că este divizat corect.
   - [x] Simulați un TTS care generează propoziția 1 în 1s, iar propoziția 2 în 0.1s. Buffer-ul **trebuie** să asigure că stream-ul final emite audio propoziției 1 complet, înainte de a livra audio-ul propoziției 2 (păstrarea secvențialității).
   - [x] Testați handler-ul HTTP pentru conexiuni întrerupte (când clientul închide conexiunea, toate goroutine-urile de sintetizare din spate se opresc prompt).

## Criterii de Acceptanță (Definition of Done)
- [x] Goroutine leak prevention activat (verificat cu tools gen `goleak`).
- [x] API-ul permite `transfer-encoding: chunked` corect pentru audio PCM.
- [x] Sistemul are o eficiență de paralelizare: dacă modelul e limitat, chunk-urile sunt calculate secvențial curat.
