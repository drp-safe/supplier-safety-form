package i18n

// translations holds all translations keyed by Lang then by message key.
// Initialized from the per-language files.
var translations = map[Lang]map[string]string{
	FR: frTranslations,
	EN: enTranslations,
	ES: esTranslations,
	DE: deTranslations,
	ZH: zhTranslations,
	JA: jaTranslations,
	VI: viTranslations,
	PT: ptTranslations,
	IT: itTranslations,
}
