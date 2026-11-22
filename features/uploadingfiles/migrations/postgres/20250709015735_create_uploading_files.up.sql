CREATE TABLE uploading_files (
	path VARCHAR(1000) PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	lease_token VARCHAR(32),
	leased_at TIMESTAMPTZ
);

-- Для оптимизации запросов.
CREATE INDEX uploading_files_lease_token_idx ON uploading_files (lease_token) WHERE lease_token IS NOT NULL;
CREATE INDEX uploading_files_created_at_leased_at_idx ON uploading_files (created_at, leased_at);

COMMENT ON TABLE uploading_files IS 'Загружаемые файлы';
COMMENT ON COLUMN uploading_files.path IS 'Путь к файлу';
COMMENT ON COLUMN uploading_files.created_at IS 'Временная метка создания';
COMMENT ON COLUMN uploading_files.lease_token IS 'Токен аренды. 16 случайных байт (128 бит) в шестнадцатеричном формате';
COMMENT ON COLUMN uploading_files.leased_at IS 'Временная метка начала аренды обработчиком';
