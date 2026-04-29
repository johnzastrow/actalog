CREATE OR REPLACE FUNCTION block_protected_users() RETURNS TRIGGER AS $$
BEGIN
    IF OLD.email = ANY(ARRAY['br8kwall@gmail.com']) THEN
        RAISE EXCEPTION 'protected user: writes blocked at db layer';
    END IF;
    IF TG_OP = 'UPDATE' THEN RETURN NEW; END IF;
    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS protected_users_no_update ON users;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION block_protected_users();

DROP TRIGGER IF EXISTS protected_users_no_delete ON users;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW EXECUTE FUNCTION block_protected_users();
