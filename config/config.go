package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Hotkey        string `json:"hotkey"`
	HotkeyRawCode uint16 `json:"hotkey_raw_code"`
	Punctuation   bool   `json:"punctuation"`
	HotwordsPath  string `json:"hotwords_path"`
	ModelID       string `json:"model_id"`
}

var DefaultConfig = Config{
	Hotkey:        "f9",
	HotkeyRawCode: 120,
	Punctuation:   true,
	ModelID:       "paraformer",
}

func LoadConfig() (Config, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return DefaultConfig, err
	}
	path := filepath.Join(configDir, "voice-writer", "config.json")

	file, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig, nil // Return default if file not exists
	}

	var cfg Config
	err = json.Unmarshal(file, &cfg)
	if err != nil {
		return DefaultConfig, err
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(configDir, "voice-writer")
	_ = os.MkdirAll(dir, 0755)
	
	path := filepath.Join(dir, "config.json")
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func GetHotwordsFilePath() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, "voice-writer", "hotwords.txt")
}

func SaveHotwords(content string) error {
	path := GetHotwordsFilePath()
	return os.WriteFile(path, []byte(content), 0644)
}

func LoadHotwords() (string, error) {
	data, err := os.ReadFile(GetHotwordsFilePath())
	if err != nil {
		return "", nil
	}
	return string(data), nil
}
