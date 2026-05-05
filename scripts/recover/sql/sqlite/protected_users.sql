DROP TRIGGER IF EXISTS protected_users_no_update;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW
WHEN OLD.email IN ('br8kwall@gmail.com')
 AND (NEW.email IS NOT OLD.email
   OR NEW.name IS NOT OLD.name
   OR NEW.role IS NOT OLD.role
   OR NEW.account_disabled IS NOT OLD.account_disabled)
BEGIN
    SELECT RAISE(ABORT, 'protected user: writes blocked at db layer');
END;

DROP TRIGGER IF EXISTS protected_users_no_delete;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW
WHEN OLD.email IN ('br8kwall@gmail.com')
BEGIN
    SELECT RAISE(ABORT, 'protected user: writes blocked at db layer');
END;
