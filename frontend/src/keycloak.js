import Keycloak from 'keycloak-js';

// Runtime env (K8s ConfigMap) takes priority over build-time env (dev).
const env = window._env_ || {};

const keycloakUrl = env.REACT_APP_KEYCLOAK_URL || import.meta.env.VITE_KEYCLOAK_URL;
const keycloakRealm = env.REACT_APP_KEYCLOAK_REALM || import.meta.env.VITE_KEYCLOAK_REALM;
const keycloakClient = env.REACT_APP_KEYCLOAK_CLIENT_ID || import.meta.env.VITE_KEYCLOAK_CLIENT_ID;

// Never throw at import time (ESM): expose the error and a null instance so
// AuthContext/api can handle a missing configuration gracefully.
let keycloak = null;
export let keycloakConfigError = null;

if (!keycloakUrl || !keycloakRealm || !keycloakClient) {
  const missing = [
    !keycloakUrl && 'REACT_APP_KEYCLOAK_URL',
    !keycloakRealm && 'REACT_APP_KEYCLOAK_REALM',
    !keycloakClient && 'REACT_APP_KEYCLOAK_CLIENT_ID',
  ].filter(Boolean).join(', ');
  keycloakConfigError = `Configuration Keycloak manquante : ${missing}. Vérifiez le ConfigMap K8s.`;
} else {
  keycloak = new Keycloak({ url: keycloakUrl, realm: keycloakRealm, clientId: keycloakClient });
}

export default keycloak;
