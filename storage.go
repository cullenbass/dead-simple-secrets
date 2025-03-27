package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log/slog"
	"os"
	"path/filepath"
)

type Storage struct {
	db     *sql.DB
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

func InitStorage(c *Config) *Storage {
	s := Storage{config: c}
	database, err := sql.Open("sqlite3", "secrets.db")
	database.SetMaxOpenConns(1)
	if err != nil {
		slog.Error("Failed to init storage", "err", err)
		os.Exit(1)
	}

	s.db = database
	rows, err := database.Query("SELECT name FROM sqlite_master WHERE type='table' AND name=?;", "secrets")
	defer rows.Close()
	if err != nil {
		slog.Error("Failed checking for DB existance", "err", err)
		os.Exit(1)
	}
	if !rows.Next() {
		s.createDatabase()
	}
	return &s
}

func (s *Storage) Cleanup() {
	s.db.Close()
}

func (s *Storage) checkOwnerExists(owner string) (exists bool) {
	err := s.db.QueryRow("SELECT (count(*) > 0) FROM secrets WHERE owner = ?", owner).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Debug("Owner does not exist in DB", "owner", owner)
		} else {
			slog.Error("Failed to query DB for owner", "owner", owner)
		}
		return false
	}
	return
}

func (s *Storage) checkSecretExists(path string) (exists bool) {
	err := s.db.QueryRow("SELECT (count(*) > 0) FROM secrets WHERE path = ?", path).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Debug("Path does not exist in DB", "path", path)
		} else {
			slog.Error("Failed to query DB for path existence", "path", path)
		}

		return false
	}
	return
}

func checkPrivsConfig(path string, config []string) bool {
	for _, configPath := range config {
		ok, err := filepath.Match(configPath, path)
		if err != nil {
			slog.Error("Failed to match paths", "path", path, "configPath", configPath, "err", err)
			continue
		}
		if ok {
			return true
		}
	}
	return false
}

func (s *Storage) checkOwnerSecret(path string, token string) bool {
	var owner string
	err := s.db.QueryRow("SELECT owner FROM secrets WHERE path = ?", path).Scan(&owner)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.Debug("Path does not exist", "path", path, "token", token)
			return false
		}
		slog.Error("Failed to query DB for owner owning secret", "path", path, "err", err)

	}
	if owner == token {
		slog.Debug("Path owned by token", "path", path, "token", token)
		return true
	}
	return false
}

func (s *Storage) checkWritePrivs(path string, token string) bool {
	if s.checkOwnerSecret(path, token) {
		slog.Debug("Token can write to path via owner", "token", token, "path", path)
		return true
	}

	for configToken, configPaths := range s.config.Write {
		if configToken == token {
			ok := checkPrivsConfig(path, configPaths)
			if ok {
				slog.Debug("Token can write to path via config", "token", token, "path", path)
				return true
			}
		}
	}
	if checkPrivsConfig(path, s.config.WriteGlobal) && !s.checkSecretExists(path) {
		slog.Debug("Token can write to path via global config", "token", token, "path", path)
		return true
	}
	return false
}

func (s *Storage) checkReadPrivs(path string, token string) bool {
	if checkPrivsConfig(path, s.config.ReadGlobal) {
		slog.Debug("Token can read via global config", "path", path, "token", token)
		return true
	}

	for configToken, configPaths := range s.config.Read {
		if configToken == token {
			if checkPrivsConfig(path, configPaths) {
				slog.Debug("Token can read via config", "path", path, "token", token)
				return true
			}
		}
	}
	ok := s.checkOwnerSecret(path, token)
	if ok {
		slog.Debug("Token can read via owner", "path", path, "token", token)
	}
	return ok
}

func (s *Storage) StoreSecret(path string, cipherText []byte, nonce []byte, token string) bool {
	if !s.checkWritePrivs(path, token) {
		slog.Debug("Cannot write secret, check privleges", "path", path, "token", token)
		return false
	}
	_, err := s.db.Exec("INSERT INTO secrets (path, owner, nonce, secret) VALUES (?, ?, ?, ?) ON CONFLICT DO UPDATE set secret=excluded.secret, nonce=excluded.nonce, owner=excluded.owner", path, token, nonce, cipherText)
	if err != nil {
		slog.Error("Failed to write secret to database", "path", path, "err", err)
		return false
	}
	return true
}

func (s *Storage) DeleteSecret(path string, token string) bool {
	if !s.checkWritePrivs(path, token) {
		slog.Info("Cannot write secret, check privleges", "path", path, "token", token)
		return false
	}
	_, err := s.db.Exec("DELETE FROM secrets WHERE path = ?", path)
	if err != nil {
		slog.Error("Failed to delete secret from database", "path", path, "err", err)
		return false
	}
	return true
}

func (s *Storage) GetSecret(path string, token string) (cipherText []byte, nonce []byte) {
	if !s.checkReadPrivs(path, token) {
		return nil, nil
	}
	rows, err := s.db.Query("SELECT secret, nonce FROM secrets WHERE path = ?", path)
	defer rows.Close()
	if err != nil {
		slog.Error("Could not query DB for secret", "path", path, "err", err)
		return nil, nil
	}
	for rows.Next() {
		err = rows.Scan(&cipherText, &nonce)
		if err != nil {
			slog.Error("Failed to scan results into return vars", "path", path, "err", err)
			return nil, nil
		}
	}
	return
}
