package config

import (
	"path/filepath"
	"strings"

	"github.com/go-faster/errors"
	"github.com/spf13/viper"

	coreerrors "github.com/kriuchkov/jiraforge/internal/core/errors"
)

const (
	ApplicationName    = "JiraForge"
	ApplicationVersion = "2.0.0"
	DefaultGeminiModel = "gemini-2.5-flash"
)

type AppInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type Settings struct {
	App       AppInfo         `json:"app"`
	Atlassian AtlassianConfig `json:"atlassian"`
	Gemini    GeminiConfig    `json:"gemini"`
}

func (s Settings) WithDefaults() Settings {
	s.Gemini = s.Gemini.WithDefaults()
	return s
}

type AtlassianConfig struct {
	Host     string `json:"host"`
	Email    string `json:"email"`
	Token    string `json:"token"`
	ProxyURL string `json:"proxy_url,omitempty"`
}

func (c AtlassianConfig) Validate() error {
	if strings.TrimSpace(c.Host) == "" {
		return coreerrors.MissingConfiguration("ATLASSIAN_HOST")
	}
	if strings.TrimSpace(c.Email) == "" {
		return coreerrors.MissingConfiguration("ATLASSIAN_EMAIL")
	}
	if strings.TrimSpace(c.Token) == "" {
		return coreerrors.MissingConfiguration("ATLASSIAN_TOKEN")
	}
	return nil
}

type GeminiConfig struct {
	APIKey string `json:"api_key,omitempty"`
	Model  string `json:"model,omitempty"`
}

func (c GeminiConfig) WithDefaults() GeminiConfig {
	c.Model = strings.TrimSpace(c.Model)
	if c.Model == "" {
		c.Model = DefaultGeminiModel
	}
	return c
}

func (c GeminiConfig) Validate() error {
	if strings.TrimSpace(c.APIKey) == "" {
		return coreerrors.MissingConfiguration("GOOGLE_API_KEY")
	}
	return nil
}

func LoadEnvFile(path string) error {
	_, err := newViper(path)
	if err != nil {
		return errors.Wrap(err, "load env file")
	}
	return nil
}

func LoadSettings() Settings {
	loader, _ := newViper("")
	return settingsFromViper(loader)
}

func LoadSettingsFromEnvFile(path string) (Settings, error) {
	loader, err := newViper(path)
	if err != nil {
		return LoadSettings(), errors.Wrap(err, "load settings from env file")
	}
	return settingsFromViper(loader), nil
}

func newViper(path string) (*viper.Viper, error) {
	loader := viper.New()
	loader.AutomaticEnv()

	trimmedPath := strings.TrimSpace(path)
	if trimmedPath == "" {
		return loader, nil
	}

	loader.SetConfigFile(trimmedPath)
	loader.SetConfigType(configTypeForPath(trimmedPath))
	if err := loader.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "read config file")
	}

	return loader, nil
}

func settingsFromViper(loader *viper.Viper) Settings {
	settings := Settings{
		App: AppInfo{
			Name:    ApplicationName,
			Version: ApplicationVersion,
		},
		Atlassian: AtlassianConfig{
			Host:     firstNonEmpty(loader.GetString("ATLASSIAN_HOST"), loader.GetString("atlassian.host")),
			Email:    firstNonEmpty(loader.GetString("ATLASSIAN_EMAIL"), loader.GetString("atlassian.email")),
			Token:    firstNonEmpty(loader.GetString("ATLASSIAN_TOKEN"), loader.GetString("atlassian.token")),
			ProxyURL: firstNonEmpty(loader.GetString("PROXY_URL"), loader.GetString("atlassian.proxy_url")),
		},
		Gemini: GeminiConfig{
			APIKey: firstNonEmpty(loader.GetString("GOOGLE_API_KEY"), loader.GetString("gemini.api_key")),
			Model:  firstNonEmpty(loader.GetString("GEMINI_MODEL"), loader.GetString("gemini.model")),
		},
	}

	return settings.WithDefaults()
}

func configTypeForPath(path string) string {
	extension := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	if extension == "" {
		return "env"
	}
	return extension
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
