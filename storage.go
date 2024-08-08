package main

import (
	"log/slog"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"path/filepath"
)

type Storage struct {
	db *sql.DB
	config *Config
}

func (s *Storage) createDatabase() {
	creationSql := `CREATE TABLE "secrets" (
		"path"	TEXT,
		"owner"	TEXT,
		"nonce"	BLOB,
		"secret"	BLOB,
		PRIMARY KEY("path")
	)`
	_, err := s.db.Exec(creationSql)
	if err != nil {
		slog.Error("Failed to create database", "err", err)
		os.Exit(1)
	}
}

func InitStorage(c *Config) *Storage{
	s := Storage{config:c}
	database, err := sql.Open("sqlite3", "secrets.db")
	database.SetMaxOpenConns(1)
	if err != nil {
		slog.Error("Failed to init storage", "err", err)
		os.Exit(1)
	}

	s.db = database
	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type='table' AND name=?;", "secrets")
	if err != nil {
		slog.Error("Failed checking for DB existance", "err", err)
		os.Exit(1)
	}
	defer rows.Close()
	if !rows.Next() {
		s.createDatabase()
	}
	return &s
}

func (s *Storage) Cleanup() {
	s.db.Close()
}

func checkPrivsConfig(path string, config []string) bool {
	for _, configPath := range config {
		ok, err := filepath.Match(configPath, path)
		if err != nil {
			slog.Info("Failed to match paths", "path", path, "configPath", configPath, "err", err)
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func (s *Storage) checkOwnerSecret(path string, owner string) bool {
	rows, err := s.db.Query("SELECT owner FROM secrets WHERE path = ?", path)
	defer rows.Close()
	if err != nil {
		slog.Error("Failed to query DB", "path", path, "err", err)
	}
	if !rows.Next() {
		return true
	}
	var currentOwner string 
	rows.Scan(&currentOwner)
	if currentOwner == owner {
		return true
	}
	return false
}

func (s *Storage) checkWritePrivs(path string, owner string) bool {
	if checkPrivsConfig(path, s.config.WriteGlobal) {
		return true
	}

	for configOwner, configPaths := range s.config.Write {
		if configOwner == owner {
			ok := checkPrivsConfig(path, configPaths)
			if ok {
				return true
			}
		}
	}

	return s.checkOwnerSecret(path, owner)
}

func (s *Storage) checkReadPrivs(path string, owner string) bool {
	if checkPrivsConfig(path, s.config.ReadGlobal) {
		return true
	}

	for configOwner, configPaths := range s.config.Read {
		if configOwner == owner {
			ok := checkPrivsConfig(path, configPaths)
			if ok {
				return true
			}
		}
	}
	return s.checkOwnerSecret(path, owner)
}

func (s *Storage) StoreSecret(path string, cipherText []byte, nonce []byte, owner string) bool {
	if !s.checkWritePrivs(path, owner) {
		slog.Info("Cannot write secret, check privleges", "path", path, "owner", owner)
		return false
	}
	_, err := s.db.Exec("INSERT INTO secrets (path, owner, nonce, secret) VALUES (?, ?, ?, ?) ON CONFLICT DO UPDATE set secret=excluded.secret, nonce=excluded.nonce, owner=excluded.owner", path, owner, nonce, cipherText)
	if err != nil {
		slog.Error("Failed to write secret to database", "path", path, "err", err)
		return false
	}
	return true
}

func (s *Storage) DeleteSecret(path string, owner string) bool {
	if !s.checkWritePrivs(path, owner) {
		slog.Info("Cannot write secret, check privleges", "path", path, "owner", owner)
		return false
	}
	_, err := s.db.Exec("DELETE FROM secrets WHERE path = ?", path)
	if err != nil {
		slog.Error("Failed to delete secret from database", "path", path, "err", err)
		return false
	}
	return true
}

func (s *Storage) GetSecret(path string , owner string) (cipherText []byte, nonce []byte) {
	if !s.checkReadPrivs(path, owner) {
		return nil, nil
	}
	rows, err := s.db.Query("SELECT secret, nonce FROM secrets WHERE path = ?", path)
	if err != nil {
		slog.Error("Could not query DB for secret", "path", path, "err", err)
		return nil, nil
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&cipherText, &nonce)
		if err != nil {
			slog.Error("Failed to scan results into return vars", "path", path, "err", err)
			return nil, nil
		}
	}
	return 
}