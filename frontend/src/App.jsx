import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './AuthContext';
import { LanguageProvider } from './LanguageContext';
import RequireAuth from './components/RequireAuth';
import Layout from './components/Layout';
import Register from './pages/Register';
import Dashboard from './pages/Dashboard';
import Admin from './pages/Admin';
import ComingSoon from './pages/ComingSoon';
import { useLanguage } from './LanguageContext';

function RequireSuperAdmin({ children }) {
  const { user } = useAuth();
  if (!user?.is_superadmin) return <Navigate to="/" replace />;
  return children;
}

function Placeholder({ titleKey }) {
  const { t } = useLanguage();
  return <ComingSoon title={t(titleKey)} />;
}

export default function App() {
  return (
    <LanguageProvider>
      <AuthProvider>
        <BrowserRouter>
          <Routes>
            <Route path="/register" element={<Register />} />
            <Route
              element={
                <RequireAuth>
                  <Layout />
                </RequireAuth>
              }
            >
              <Route path="/" element={<Dashboard />} />
              <Route path="/library" element={<Placeholder titleKey="nav.library" />} />
              <Route path="/proposals" element={<Placeholder titleKey="nav.proposals" />} />
              <Route path="/forms" element={<Placeholder titleKey="nav.myForms" />} />
              <Route path="/editor" element={<Placeholder titleKey="nav.editor" />} />
              <Route path="/solicitations" element={<Placeholder titleKey="nav.solicitations" />} />
              <Route path="/results" element={<Placeholder titleKey="nav.results" />} />
              <Route
                path="/admin"
                element={
                  <RequireSuperAdmin>
                    <Admin />
                  </RequireSuperAdmin>
                }
              />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </AuthProvider>
    </LanguageProvider>
  );
}
