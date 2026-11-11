import axios from 'axios';
import keycloak from './keycloak';

const api = axios.create({ baseURL: '/api' });

// Inject the Keycloak Bearer token on every request, and the current UI
// language so the backend can answer error messages already translated.
api.interceptors.request.use(async (config) => {
  config.headers['Accept-Language'] = localStorage.getItem('app_language') || 'fr';

  if (import.meta.env.MODE === 'test' || window.__MOCK_AUTH__ || !keycloak) {
    return config;
  }
  try {
    await keycloak.updateToken(30);
  } catch {
    // Token refresh failed — Keycloak will handle re-login.
  }
  if (keycloak.token) {
    config.headers.Authorization = `Bearer ${keycloak.token}`;
  }
  return config;
});

export const meAPI = {
  get: () => api.get('/me'),
};

export const registrationAPI = {
  register: (data) => api.post('/register', data),
};

export const referenceAPI = {
  listSectors: () => api.get('/sectors'),
  listThemes: () => api.get('/themes'),
};

export const adminAPI = {
  getEmailConfig: () => api.get('/admin/email-config'),
  updateEmailConfig: (data) => api.put('/admin/email-config', data),
  testEmailConfig: (testEmail) => api.post('/admin/email-config/test', { test_email: testEmail }),

  listSectors: () => api.get('/admin/sectors'),
  createSector: (data) => api.post('/admin/sectors', data),
  updateSector: (code, data) => api.put(`/admin/sectors/${code}`, data),
  deleteSector: (code) => api.delete(`/admin/sectors/${code}`),

  getOAuthConfig: () => api.get('/admin/oauth-config'),
  getDBInfo: () => api.get('/admin/db-info'),
  listMigrations: () => api.get('/admin/migrations'),
};

export default api;
