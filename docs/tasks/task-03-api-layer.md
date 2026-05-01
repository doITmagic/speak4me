# Task 3: Baza de Date și Input / API Layer

## Obiectiv Principal
Orice aplicație necesită o memorie persistentă și un mod de comunicare. Acest task asigură inițializarea SQLite pentru salvarea stării (preferințele de voce) și un server HTTP (API) care preia comenzi de la Wails (frontend) sau de la un client REST extern.

## Pre-rechizite
- Pachetele `core` (Task 1) și `mock` (Task 2) sunt finalizate și testate.
- Librăria `github.com/mattn/go-sqlite3` instalată (cu CGO activat, sau o alternativă pură Go precum `modernc.org/sqlite` dacă se dorește compilare cross-platform ușoară). Se va prefera **pure Go SQLite** (ex. `github.com/glebarez/go-sqlite`) pentru build-uri cross-platform simple, așa cum e uzual în desktop apps.

## Pași de Execuție (Step-by-Step)

### Pasul 3.1: Stocarea (Database Layer)
1. Creați `pkg/db/sqlite.go`.
2. Scrieți o funcție `InitDB(dbPath string) (*sql.DB, error)` care va crea fișierul `speak4me.db` (ex. în `~/.config/doitmagic/speak4me/` sau un director temporar pentru teste).
3. Structura bazei de date (Schema):
   - Tabel `settings` (key TEXT PRIMARY KEY, value TEXT).
   - Acesta va reține, de exemplu: `(key='last_used_voice', value='xtts_ro_1')`.

### Pasul 3.2: API Server (HTTP/REST)
1. Creați `pkg/api/server.go` și `pkg/api/handlers.go`.
2. Se va folosi ruter-ul HTTP nativ din Go (`net/http.NewServeMux()` - Go 1.22 support).
3. **Structura Handler-elor:**
   - Un struct `APIServer` care ține o referință către `ModelRegistry` și către `*sql.DB`.
   - `GET /api/v1/voices`: returnează JSON cu rezultatul de la `registry.ListAllVoices()`.
   - `POST /api/v1/synthesize`: acceptă un JSON cu structura `SynthesisRequest`. Validează JSON-ul, apelează `registry.GetProvider()`, apelează `Synthesize()`, și returnează payload-ul binar (ex. `Content-Type: audio/wav`). Dacă apare o eroare, returnează 400 Bad Request sau 500 Internal Error.

### Pasul 3.3: Testarea Integrării HTTP și DB (BDD)
1. Generați suita Ginkgo în `pkg/api/`.
2. Pentru teste, folosiți baza de date in-memory SQLite: `InitDB(":memory:")`.
3. Instanțiați API-ul injectând un `ModelRegistry` ce conține un `MockTTS`.
4. Folosiți `httptest.NewRecorder()` pentru a testa handlerele:
   - [x] Un GET pe `/voices` întoarce status 200 și un array JSON cu vocea din mock.
   - [x] Un POST valid pe `/synthesize` cu text și voce corectă întoarce 200 OK și mime-type audio.
   - [x] Un POST cu o voce inexistentă întoarce 400/404 cu mesaj JSON de eroare predictibil.

## Criterii de Acceptanță (Definition of Done)
- [x] Fișierele SQLite și rutele funcționează perfect cap la cap.
- [x] Testele acoperă serializarea/deserializarea JSON și gestionarea erorilor HTTP.
- [x] Coverage rezonabil pe `handlers.go`.
