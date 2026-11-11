import React, { useEffect } from 'react';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';

// Routes under this guard are the ones that actually need a session — unlike
// MailingTools-Frontend (always behind Keycloak), this app also serves a
// public /register page, so login is enforced per-route rather than globally
// (see AuthContext's check-sso init).
export default function RequireAuth({ children }) {
  const { authenticated, login } = useAuth();
  const { t } = useLanguage();

  useEffect(() => {
    if (!authenticated) login();
  }, [authenticated, login]);

  if (!authenticated) {
    return <div className="flex items-center justify-center h-screen text-neutral-500">{t('common.loading')}</div>;
  }
  return children;
}
