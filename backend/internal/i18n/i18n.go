package i18n

import (
	"fmt"
	"net/http"
	"strings"
)

// Lang represents a supported language
type Lang string

const (
	FR Lang = "fr"
	EN Lang = "en"
	ES Lang = "es"
	DE Lang = "de"
	ZH Lang = "zh"
	JA Lang = "ja"
	VI Lang = "vi"
	PT Lang = "pt"
	IT Lang = "it"
)

// T returns the translation for the given key in the given language.
// Falls back to FR, then returns the key itself.
func T(lang Lang, key string) string {
	if msgs, ok := translations[lang]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	if msgs, ok := translations[FR]; ok {
		if msg, ok := msgs[key]; ok {
			return msg
		}
	}
	return key
}

// Tf returns a formatted translation (like fmt.Sprintf)
func Tf(lang Lang, key string, args ...interface{}) string {
	return fmt.Sprintf(T(lang, key), args...)
}

// ParseLang parses a language string, defaulting to FR
func ParseLang(s string) Lang {
	s = strings.ToLower(strings.TrimSpace(s))
	if strings.HasPrefix(s, "en") {
		return EN
	}
	if strings.HasPrefix(s, "es") {
		return ES
	}
	if strings.HasPrefix(s, "de") {
		return DE
	}
	if strings.HasPrefix(s, "zh") {
		return ZH
	}
	if strings.HasPrefix(s, "ja") {
		return JA
	}
	if strings.HasPrefix(s, "vi") {
		return VI
	}
	if strings.HasPrefix(s, "pt") {
		return PT
	}
	if strings.HasPrefix(s, "it") {
		return IT
	}
	return FR
}

// LangFromRequest extracts the language from the Accept-Language header.
func LangFromRequest(r *http.Request) Lang {
	header := r.Header.Get("Accept-Language")
	if header == "" {
		return FR
	}
	return ParseLang(header)
}

// keycloakLocales : correspondance entre nos codes et ceux sous lesquels
// Keycloak livre SES propres traductions (cf. pra2/backend/internal/i18n :
// Keycloak nomme le chinois "zh-CN", pas "zh").
var keycloakLocales = map[Lang]string{
	ZH: "zh-CN",
}

// KeycloakLocale rend l'étiquette à poser dans l'attribut "locale" d'un
// compte Keycloak (langue de l'écran de connexion et de l'email "définissez
// votre mot de passe"). La valeur doit figurer dans les supportedLocales du
// realm (k8s/keycloak/base/realm-supplier-safety-form.json).
func KeycloakLocale(lang Lang) string {
	if tag, ok := keycloakLocales[lang]; ok {
		return tag
	}
	return string(lang)
}
