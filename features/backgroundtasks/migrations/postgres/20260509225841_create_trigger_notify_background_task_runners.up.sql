/*
notify_background_task_runners уведомляет один раз каждый runner из фоновых задач в запросе.
*/
CREATE FUNCTION notify_background_task_runners() RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
	DECLARE
		runner_name TEXT;
	BEGIN
		FOR runner_name IN
			SELECT DISTINCT runner
			FROM tasks
		LOOP
			PERFORM pg_notify('background_tasks', runner_name);
		END LOOP;
		RETURN NULL;
	END;
$$;

CREATE TRIGGER trigger_notify_background_task_runners
AFTER INSERT ON background_tasks
REFERENCING NEW TABLE tasks
FOR EACH STATEMENT
EXECUTE FUNCTION notify_background_task_runners();
