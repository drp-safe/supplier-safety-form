-- Une seule ligne, comme pra2 : paramétrage du serveur mail, modifiable depuis
-- la section Administration (mot de passe chiffré au repos par le backend).
CREATE TABLE email_config (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    smtp_host       TEXT NOT NULL DEFAULT '',
    smtp_port       INTEGER NOT NULL DEFAULT 587,
    smtp_username   TEXT NOT NULL DEFAULT '',
    smtp_password   TEXT NOT NULL DEFAULT '',
    smtp_encryption TEXT NOT NULL DEFAULT 'tls',
    from_address    TEXT NOT NULL DEFAULT '',
    from_name       TEXT NOT NULL DEFAULT '',
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO email_config (from_name) VALUES ('Supplier Safety Form');
