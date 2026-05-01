package db

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
)

// InitDB inițializează conexiunea la baza de date SQLite și creează tabelele necesare
func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("eroare la deschiderea bazei de date: %w", err)
	}

	// Crearea tabelului settings
	query := `
	CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);
	`
	if _, err := db.Exec(query); err != nil {
		return nil, fmt.Errorf("eroare la crearea tabelului settings: %w", err)
	}

	return db, nil
}

// SetSetting salvează sau actualizează o setare
func SetSetting(db *sql.DB, key, value string) error {
	query := `INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`
	_, err := db.Exec(query, key, value)
	return err
}

// GetSetting recuperează o setare după cheie
func GetSetting(db *sql.DB, key string) (string, error) {
	var value string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Returnează empty dacă nu există, fără eroare fatală
		}
		return "", err
	}
	return value, nil
}
