#!/bin/sh
# Generate env-config.js from REACT_APP_* environment variables at runtime, so
# the same Docker image is configured per-environment via a K8s ConfigMap.

cat <<EOF > /app/build/env-config.js
window._env_ = {
  REACT_APP_API_URL: "${REACT_APP_API_URL}",
  REACT_APP_KEYCLOAK_URL: "${REACT_APP_KEYCLOAK_URL}",
  REACT_APP_KEYCLOAK_REALM: "${REACT_APP_KEYCLOAK_REALM}",
  REACT_APP_KEYCLOAK_CLIENT_ID: "${REACT_APP_KEYCLOAK_CLIENT_ID}",
  APP_ENV: "${APP_ENV}",
  REACT_APP_VERSION: "${REACT_APP_VERSION}",
};
EOF

exec "$@"
