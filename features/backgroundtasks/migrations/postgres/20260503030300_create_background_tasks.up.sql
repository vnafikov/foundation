CREATE TYPE background_task_runner_type AS ENUM ('compute', 'filesystem', 'network', 'heavy');

CREATE TABLE background_tasks (
	id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name VARCHAR(100) NOT NULL,
	runner background_task_runner_type NOT NULL,
	ordering_key VARCHAR(200),
	deduplication_key VARCHAR(200),
	payload JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
	leased_by_worker_id INTEGER,
	leased_at TIMESTAMPTZ,

	CONSTRAINT background_tasks_ordering_key_required_for_deduplication_check CHECK (
		deduplication_key IS NULL
		OR ordering_key IS NOT NULL
	),
	CONSTRAINT background_tasks_deduplication_cleared_on_lease_check CHECK (
		leased_at IS NULL
		OR deduplication_key IS NULL
	),
	CONSTRAINT background_tasks_lease_check CHECK (
		(leased_by_worker_id IS NULL) = (leased_at IS NULL)
	)
);

CREATE UNIQUE INDEX background_tasks_runner_deduplication_key_idx ON background_tasks (runner, deduplication_key)
WHERE deduplication_key IS NOT NULL;

-- Для оптимизации запросов.
CREATE INDEX background_tasks_runner_idx ON background_tasks (runner) WHERE leased_at IS NOT NULL;
CREATE INDEX background_tasks_runner_ordering_key_id_idx ON background_tasks (runner, ordering_key, id) WHERE ordering_key IS NOT NULL;
CREATE INDEX background_tasks_leased_by_worker_id_idx ON background_tasks (leased_by_worker_id) WHERE leased_by_worker_id IS NOT NULL;
CREATE INDEX background_tasks_runner_id_idx ON background_tasks (runner, id);
CREATE INDEX background_tasks_runner_available_at_id_idx ON background_tasks (runner, available_at, id);

COMMENT ON TABLE background_tasks IS 'Фоновые задачи';
COMMENT ON COLUMN background_tasks.name IS 'Название';
COMMENT ON COLUMN background_tasks.runner IS 'Исполнитель. heavy для тяжёлых фоновых задач, которые не должны влиять на работу с приложением';
COMMENT ON COLUMN background_tasks.ordering_key IS 'Ключ последовательного выполнения. Фоновые задачи одного runner с одинаковым ключом выполняются в порядке id';
COMMENT ON COLUMN background_tasks.deduplication_key IS 'Ключ дедупликации ожидающих фоновых задач одного исполнителя';
COMMENT ON COLUMN background_tasks.payload IS 'Данные';
COMMENT ON COLUMN background_tasks.created_at IS 'Временная метка создания';
COMMENT ON COLUMN background_tasks.available_at IS 'Временная метка начала разрешения на аренду';
COMMENT ON COLUMN background_tasks.leased_by_worker_id IS 'ID обработчика, который взял в работу';
COMMENT ON COLUMN background_tasks.leased_at IS 'Временная метка начала аренды обработчиком';
