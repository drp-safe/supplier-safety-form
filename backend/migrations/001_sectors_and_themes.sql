-- Listes de référence, gérées en CRUD par la section Administration (pas des
-- valeurs figées en dur) : secteurs d'activité des organisations inscrites, et
-- thèmes qui classent les questions de la bibliothèque.

CREATE TABLE sectors (
    code          TEXT PRIMARY KEY,
    label         TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    active        BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO sectors (code, label, display_order) VALUES
    ('services_numeriques_esn',  'Services numériques & ESN',   10),
    ('industrie_manufacturiere', 'Industrie manufacturière',    20),
    ('transport_logistique',     'Transport & logistique',      30),
    ('commerce_distribution',    'Commerce & distribution',     40),
    ('sante_medico_social',      'Santé & médico-social',       50),
    ('finance_assurance',        'Finance & assurance',         60),
    ('administration_publique',  'Administration publique',     70),
    ('energie_utilities',        'Énergie & utilities',         80),
    ('construction_btp',         'Construction & BTP',          90),
    ('agroalimentaire',          'Agroalimentaire',            100),
    ('education_recherche',      'Éducation & recherche',      110),
    ('autre',                    'Autre',                      999);

CREATE TABLE themes (
    code          TEXT PRIMARY KEY,
    label         TEXT NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0
);

INSERT INTO themes (code, label, display_order) VALUES
    ('gouvernance',             'Gouvernance',               10),
    ('gestion_acces',           'Gestion des accès',         20),
    ('sauvegarde_continuite',   'Sauvegarde & continuité',   30),
    ('securite_reseau',         'Sécurité réseau',           40),
    ('protection_donnees',      'Protection des données',    50),
    ('gestion_incidents',       'Gestion des incidents',     60),
    ('developpement_securise',  'Développement sécurisé',    70),
    ('sensibilisation',         'Sensibilisation',           80),
    ('securite_physique',       'Sécurité physique',         90);
