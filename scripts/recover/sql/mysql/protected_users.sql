DROP TRIGGER IF EXISTS protected_users_no_update;
CREATE TRIGGER protected_users_no_update
BEFORE UPDATE ON users
FOR EACH ROW
BEGIN
    IF OLD.email = 'br8kwall@gmail.com' THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'protected user: writes blocked at db layer';
    END IF;
END;

DROP TRIGGER IF EXISTS protected_users_no_delete;
CREATE TRIGGER protected_users_no_delete
BEFORE DELETE ON users
FOR EACH ROW
BEGIN
    IF OLD.email = 'br8kwall@gmail.com' THEN
        SIGNAL SQLSTATE '45000'
            SET MESSAGE_TEXT = 'protected user: writes blocked at db layer';
    END IF;
END;
