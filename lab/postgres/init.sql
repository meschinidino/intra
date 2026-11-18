-- Intra lab database. Fake data only.
CREATE TABLE IF NOT EXISTS secrets (id int PRIMARY KEY, value text NOT NULL);
INSERT INTO secrets VALUES (1, 'CANARY-SECRET-123') ON CONFLICT (id) DO NOTHING;

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'intra_broker_admin') THEN
    CREATE ROLE intra_broker_admin LOGIN;
  END IF;
END $$;

-- The broker admin gets no data privileges. It may only call these functions.
REVOKE ALL ON secrets FROM intra_broker_admin;

CREATE OR REPLACE FUNCTION intra_create_grant_role(p_role text, p_password text, p_valid_until timestamptz, p_objects text[] DEFAULT ARRAY['secrets'])
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  IF p_role !~ '^intra_g_[a-z0-9_]+$' THEN
    RAISE EXCEPTION 'bad grant role';
  END IF;
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = p_role) THEN
    EXECUTE format('ALTER ROLE %I WITH LOGIN PASSWORD %L VALID UNTIL %L', p_role, p_password, p_valid_until);
  ELSE
    EXECUTE format('CREATE ROLE %I WITH LOGIN PASSWORD %L VALID UNTIL %L', p_role, p_password, p_valid_until);
  END IF;
  EXECUTE format('ALTER ROLE %I SET default_transaction_read_only = on', p_role);
  EXECUTE format('GRANT CONNECT ON DATABASE %I TO %I', current_database(), p_role);
  EXECUTE format('GRANT USAGE ON SCHEMA public TO %I', p_role);
  EXECUTE format('REVOKE ALL ON ALL TABLES IN SCHEMA public FROM %I', p_role);
  FOR i IN 1 .. coalesce(array_length(p_objects, 1), 0) LOOP
    IF p_objects[i] !~ '^[a-z_][a-z0-9_]*$' THEN
      RAISE EXCEPTION 'bad grant object';
    END IF;
    EXECUTE format('GRANT SELECT ON %I TO %I', p_objects[i], p_role);
  END LOOP;
END;
$$;

CREATE OR REPLACE FUNCTION intra_drop_grant_role(p_role text)
RETURNS void
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = pg_catalog, public
AS $$
BEGIN
  IF p_role !~ '^intra_g_[a-z0-9_]+$' THEN
    RAISE EXCEPTION 'bad grant role';
  END IF;
  PERFORM pg_terminate_backend(pid)
    FROM pg_stat_activity
   WHERE usename = p_role AND pid <> pg_backend_pid();
  IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname = p_role) THEN
    EXECUTE format('ALTER ROLE %I WITH NOLOGIN', p_role);
    EXECUTE format('DROP OWNED BY %I', p_role);
    EXECUTE format('DROP ROLE %I', p_role);
  END IF;
END;
$$;

REVOKE ALL ON FUNCTION intra_create_grant_role(text, text, timestamptz, text[]) FROM PUBLIC;
REVOKE ALL ON FUNCTION intra_drop_grant_role(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION intra_create_grant_role(text, text, timestamptz, text[]) TO intra_broker_admin;
GRANT EXECUTE ON FUNCTION intra_drop_grant_role(text) TO intra_broker_admin;
