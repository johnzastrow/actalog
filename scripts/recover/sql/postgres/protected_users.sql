CREATE OR REPLACE FUNCTION block_protected_users() RETURNS TRIGGER AS $$
BEGIN
    IF NOT (OLD.email = ANY(ARRAY['br8kwall@gmail.com'])) THEN
        IF TG_OP = 'UPDATE' THEN RETURN NEW; END IF;
        RETURN OLD;
    END IF;
    IF TG_OP = 'DELETE' THEN
        RAISE EXCEPTION 'protected user: writes blocked at db layer';
    END IF;
    IF NEW.email IS DISTINCT FROM OLD.email
       OR NEW.name IS DISTINCT FROM OLD.name
       OR NEW.role IS DISTINCT FROM OLD.role
       OR NEW.account_disabled IS DISTINCT FROM OLD.account_disabled THEN
        RAISE EXCEPTION 'protected user: writes blocked at db layer';
    END IF;
    RETURN NEW;
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
