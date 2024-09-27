-- FUNCTIONS
DO $$
BEGIN
	CREATE FUNCTION article_management.update_updated_at_column( )
		RETURNS TRIGGER
		LANGUAGE plpgsql
		AS $ func_body $
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
END;
	$func_body$;
	EXCEPTION
	WHEN duplicate_function THEN
		NULL;
END;

$$;
