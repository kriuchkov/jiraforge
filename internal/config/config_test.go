package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettingsAppliesDefaults(t *testing.T) {
	t.Setenv("ATLASSIAN_HOST", "https://example.atlassian.net")
	t.Setenv("ATLASSIAN_EMAIL", "user@example.com")
	t.Setenv("ATLASSIAN_TOKEN", "token")
	t.Setenv("GOOGLE_API_KEY", "key")
	t.Setenv("GEMINI_MODEL", "")

	settings := LoadSettings()

	if settings.App.Name != ApplicationName {
		t.Fatalf("unexpected app name: got %q want %q", settings.App.Name, ApplicationName)
	}
	if settings.App.Version != ApplicationVersion {
		t.Fatalf("unexpected app version: got %q want %q", settings.App.Version, ApplicationVersion)
	}
	if settings.Atlassian.Host != "https://example.atlassian.net" {
		t.Fatalf("unexpected host: got %q", settings.Atlassian.Host)
	}
	if settings.Gemini.Model != DefaultGeminiModel {
		t.Fatalf("unexpected model: got %q want %q", settings.Gemini.Model, DefaultGeminiModel)
	}
}

func TestLoadSettingsFromEnvFileEmptyPath(t *testing.T) {
	t.Setenv("ATLASSIAN_HOST", "https://example.atlassian.net")
	t.Setenv("ATLASSIAN_EMAIL", "user@example.com")
	t.Setenv("ATLASSIAN_TOKEN", "token")

	settings, err := LoadSettingsFromEnvFile("")
	if err != nil {
		t.Fatalf("LoadSettingsFromEnvFile returned error: %v", err)
	}
	if settings.Atlassian.Email != "user@example.com" {
		t.Fatalf("unexpected email: got %q", settings.Atlassian.Email)
	}
}

func TestLoadSettingsFromEnvFileLoadsValues(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := "ATLASSIAN_HOST=https://file.atlassian.net\n" +
		"ATLASSIAN_EMAIL=file@example.com\n" +
		"ATLASSIAN_TOKEN=file-token\n" +
		"GOOGLE_API_KEY=file-key\n" +
		"GEMINI_MODEL=gemini-2.5-pro\n"

	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	settings, err := LoadSettingsFromEnvFile(envFile)
	if err != nil {
		t.Fatalf("LoadSettingsFromEnvFile returned error: %v", err)
	}
	if settings.Atlassian.Host != "https://file.atlassian.net" {
		t.Fatalf("unexpected host: got %q", settings.Atlassian.Host)
	}
	if settings.Atlassian.Email != "file@example.com" {
		t.Fatalf("unexpected email: got %q", settings.Atlassian.Email)
	}
	if settings.Gemini.APIKey != "file-key" {
		t.Fatalf("unexpected api key: got %q", settings.Gemini.APIKey)
	}
	if settings.Gemini.Model != "gemini-2.5-pro" {
		t.Fatalf("unexpected model: got %q", settings.Gemini.Model)
	}
}

func TestLoadSettingsFromEnvFileEnvOverridesFile(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	content := "ATLASSIAN_HOST=https://file.atlassian.net\n" +
		"ATLASSIAN_EMAIL=file@example.com\n" +
		"ATLASSIAN_TOKEN=file-token\n" +
		"GOOGLE_API_KEY=file-key\n" +
		"GEMINI_MODEL=gemini-2.5-pro\n"

	if err := os.WriteFile(envFile, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write env file: %v", err)
	}

	t.Setenv("ATLASSIAN_EMAIL", "env@example.com")
	t.Setenv("GEMINI_MODEL", "gemini-2.5-flash-lite")

	settings, err := LoadSettingsFromEnvFile(envFile)
	if err != nil {
		t.Fatalf("LoadSettingsFromEnvFile returned error: %v", err)
	}
	if settings.Atlassian.Host != "https://file.atlassian.net" {
		t.Fatalf("unexpected host: got %q", settings.Atlassian.Host)
	}
	if settings.Atlassian.Email != "env@example.com" {
		t.Fatalf("unexpected email: got %q", settings.Atlassian.Email)
	}
	if settings.Gemini.Model != "gemini-2.5-flash-lite" {
		t.Fatalf("unexpected model: got %q", settings.Gemini.Model)
	}
}
