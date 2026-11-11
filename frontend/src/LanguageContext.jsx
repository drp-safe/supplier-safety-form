import React, { createContext, useContext, useState, useCallback, useEffect } from 'react';
// French stays statically bundled: it is both the default language and the
// guaranteed fallback used by t() for any missing key. The other locales are
// code-split and fetched on demand (see loadLocale) so a user only downloads
// the language they actually use. Same convention as pra2's LanguageContext.
import fr from './locales/fr.json';

const localeLoaders = {
  en: () => import('./locales/en.json'),
  es: () => import('./locales/es.json'),
  de: () => import('./locales/de.json'),
  zh: () => import('./locales/zh.json'),
  ja: () => import('./locales/ja.json'),
  vi: () => import('./locales/vi.json'),
  pt: () => import('./locales/pt.json'),
  it: () => import('./locales/it.json'),
};

export const SUPPORTED_LANGUAGES = ['fr', 'en', 'es', 'de', 'zh', 'ja', 'vi', 'pt', 'it'];
const isSupported = (lang) => SUPPORTED_LANGUAGES.includes(lang);

export const languageFlags = { fr: '🇫🇷', en: '🇬🇧', de: '🇩🇪', es: '🇪🇸', zh: '🇨🇳', ja: '🇯🇵', vi: '🇻🇳', pt: '🇵🇹', it: '🇮🇹' };
export const languageNames = { fr: 'Français', en: 'English', de: 'Deutsch', es: 'Español', zh: '中文', ja: '日本語', vi: 'Tiếng Việt', pt: 'Português', it: 'Italiano' };

const LanguageContext = createContext();

export function LanguageProvider({ children }) {
  const [language, setLanguage] = useState(() => {
    const stored = localStorage.getItem('app_language');
    return stored && isSupported(stored) ? stored : 'fr';
  });
  const [loaded, setLoaded] = useState({ fr });

  const loadLocale = useCallback((lang) => {
    if (!localeLoaders[lang]) return Promise.resolve(); // fr or unknown
    return localeLoaders[lang]()
      .then((mod) => {
        const data = mod.default || mod;
        setLoaded((prev) => (prev[lang] ? prev : { ...prev, [lang]: data }));
      })
      .catch(() => {
        // Chunk failed to load — t() keeps falling back to French.
      });
  }, []);

  useEffect(() => {
    if (language !== 'fr') loadLocale(language);
  }, [language, loadLocale]);

  const changeLanguage = (lang) => {
    if (!isSupported(lang)) return;
    loadLocale(lang).finally(() => {
      setLanguage(lang);
      localStorage.setItem('app_language', lang);
    });
  };

  const t = (key, vars) => {
    const keys = key.split('.');
    let value = loaded[language] || loaded.fr;
    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k];
      } else {
        let fallback = loaded.fr;
        for (const fk of keys) {
          if (fallback && typeof fallback === 'object' && fk in fallback) {
            fallback = fallback[fk];
          } else {
            return key;
          }
        }
        value = fallback;
        break;
      }
    }
    if (vars && typeof value === 'string') {
      return value.replace(/\{\{(\w+)\}\}/g, (_, k) => (vars[k] ?? `{{${k}}}`));
    }
    return value;
  };

  return (
    <LanguageContext.Provider value={{ language, changeLanguage, t, availableLanguages: SUPPORTED_LANGUAGES, languageFlags, languageNames }}>
      {children}
    </LanguageContext.Provider>
  );
}

export function useLanguage() {
  const context = useContext(LanguageContext);
  if (!context) {
    throw new Error('useLanguage must be used within a LanguageProvider');
  }
  return context;
}
