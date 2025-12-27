package services

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// I18n provides internationalization support
// TEMPLATE.md PART 29 - NON-NEGOTIABLE
type I18n struct {
	mu            sync.RWMutex
	// lang -> key -> translation
	translations  map[string]map[string]string
	defaultLang   string
	supportedLang []string
}

// NewI18n creates a new internationalization service
func NewI18n(fs embed.FS, defaultLang string) (*I18n, error) {
	i18n := &I18n{
		translations:  make(map[string]map[string]string),
		defaultLang:   defaultLang,
		supportedLang: []string{},
	}

	// Load all locale files from embedded FS
	entries, err := fs.ReadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("failed to read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		lang := strings.TrimSuffix(entry.Name(), ".json")
		data, err := fs.ReadFile(fmt.Sprintf("locales/%s", entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read locale file %s: %w", entry.Name(), err)
		}

		var translations map[string]string
		if err := json.Unmarshal(data, &translations); err != nil {
			return nil, fmt.Errorf("failed to parse locale file %s: %w", entry.Name(), err)
		}

		i18n.translations[lang] = translations
		i18n.supportedLang = append(i18n.supportedLang, lang)
	}

	return i18n, nil
}

// T translates a key for the given language
// Falls back to default language if key not found
func (i *I18n) T(lang, key string) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Try requested language
	if translations, ok := i.translations[lang]; ok {
		if text, ok := translations[key]; ok {
			return text
		}
	}

	// Fall back to default language
	if translations, ok := i.translations[i.defaultLang]; ok {
		if text, ok := translations[key]; ok {
			return text
		}
	}

	// Return key if translation not found (helpful for debugging)
	return fmt.Sprintf("[%s]", key)
}

// ParseAcceptLanguage parses the Accept-Language header and returns the best match
// Header format: "en-US,en;q=0.9,es;q=0.8"
func (i *I18n) ParseAcceptLanguage(header string) string {
	if header == "" {
		return i.defaultLang
	}

	// Split by comma
	languages := strings.Split(header, ",")
	bestMatch := i.defaultLang
	highestQ := 0.0

	for _, lang := range languages {
		// Parse language and quality
		parts := strings.Split(strings.TrimSpace(lang), ";")
		langCode := strings.ToLower(strings.TrimSpace(parts[0]))

		// Extract base language (en-US -> en)
		if idx := strings.Index(langCode, "-"); idx > 0 {
			langCode = langCode[:idx]
		}

		// Parse quality factor (default 1.0)
		q := 1.0
		if len(parts) > 1 {
			if strings.HasPrefix(parts[1], "q=") {
				fmt.Sscanf(parts[1], "q=%f", &q)
			}
		}

		// Check if we support this language
		if i.isSupported(langCode) && q > highestQ {
			bestMatch = langCode
			highestQ = q
		}
	}

	return bestMatch
}

// isSupported checks if a language is supported
func (i *I18n) isSupported(lang string) bool {
	i.mu.RLock()
	defer i.mu.RUnlock()

	_, ok := i.translations[lang]
	return ok
}

// GetSupportedLanguages returns list of supported languages
func (i *I18n) GetSupportedLanguages() []string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	langs := make([]string, len(i.supportedLang))
	copy(langs, i.supportedLang)
	return langs
}

// GetDefaultLanguage returns the default language
func (i *I18n) GetDefaultLanguage() string {
	return i.defaultLang
}
