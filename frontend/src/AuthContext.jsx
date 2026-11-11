import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import { meAPI } from './api';
import keycloak, { keycloakConfigError } from './keycloak';

const AuthContext = createContext(null);

// Contrairement à MailingTools-Frontend (outil interne, toujours derrière
// Keycloak), ce SaaS a une page d'inscription publique (/register) : on ne
// peut donc pas forcer la connexion dès le chargement de l'application
// (onLoad: 'login-required'). On vérifie juste silencieusement s'il existe
// déjà une session (check-sso) et on expose authenticated/login ; c'est aux
// routes protégées d'appeler login() si besoin (voir components/RequireAuth).
export function AuthProvider({ children }) {
  const [authenticated, setAuthenticated] = useState(false);
  const [loading, setLoading] = useState(!keycloakConfigError);
  const [user, setUser] = useState(null);
  const didInit = useRef(false);

  const buildUser = useCallback(async (tokenParsed) => {
    const base = {
      id: tokenParsed.sub,
      email: tokenParsed.email,
      name: tokenParsed.name || `${tokenParsed.given_name || ''} ${tokenParsed.family_name || ''}`.trim(),
    };
    try {
      const res = await meAPI.get();
      if (res.data) Object.assign(base, res.data);
    } catch {
      // fall back to token values
    }
    return base;
  }, []);

  useEffect(() => {
    if (didInit.current) return;
    didInit.current = true;

    if (import.meta.env.MODE === 'test' || window.__MOCK_AUTH__) {
      setAuthenticated(true);
      setUser({ id: 'test', email: 'test@example.com', name: 'Test User', is_superadmin: true });
      setLoading(false);
      return;
    }

    if (keycloakConfigError || !keycloak) return;

    keycloak.init({
      onLoad: 'check-sso',
      silentCheckSsoRedirectUri: window.location.origin + '/silent-check-sso.html',
      pkceMethod: 'S256',
      checkLoginIframe: false,
    })
      .then(async (auth) => {
        setAuthenticated(auth);
        if (auth && keycloak.tokenParsed) {
          setUser(await buildUser(keycloak.tokenParsed));
        }
        setLoading(false);
      })
      .catch((err) => {
        console.error('Keycloak init failed:', err);
        setLoading(false);
      });

    keycloak.onTokenExpired = () => {
      keycloak.updateToken(30).catch(() => keycloak.login());
    };
  }, [buildUser]);

  const login = useCallback((redirectUri) => {
    keycloak.login({ redirectUri: redirectUri || window.location.href });
  }, []);

  const logout = useCallback(() => {
    keycloak.logout({ redirectUri: window.location.origin + '/' });
  }, []);

  if (keycloakConfigError && !window.__MOCK_AUTH__) {
    return (
      <div className="flex flex-col items-center justify-center h-screen text-center px-6">
        <h2 className="text-lg font-semibold text-brand-navy mb-2">Erreur de configuration</h2>
        <p className="text-danger max-w-md">{keycloakConfigError}</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen text-neutral-500">Chargement…</div>
    );
  }

  return (
    <AuthContext.Provider value={{ authenticated, user, login, logout, keycloak }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within an AuthProvider');
  return ctx;
}
