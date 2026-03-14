-- +goose Up
CREATE OR REPLACE FUNCTION log_item_changes() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        INSERT INTO items_history (item_id, action, old_data, new_data, changed_by)
        VALUES (NEW.id, 'INSERT', NULL, row_to_json(NEW)::jsonb, COALESCE(current_setting('warehouse_control.changed_by', TRUE), 'unknown'));
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO items_history (item_id, action, old_data, new_data, changed_by)
        VALUES (NEW.id, 'UPDATE', row_to_json(OLD)::jsonb, row_to_json(NEW)::jsonb, COALESCE(current_setting('warehouse_control.changed_by', TRUE), 'unknown'));
    ELSIF (TG_OP = 'DELETE') THEN
        INSERT INTO items_history (item_id, action, old_data, new_data, changed_by)
        VALUES (OLD.id, 'DELETE', row_to_json(OLD)::jsonb, NULL, COALESCE(current_setting('warehouse_control.changed_by', TRUE), 'unknown'));
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER items_history_trigger
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW EXECUTE PROCEDURE log_item_changes();

-- +goose Down
DROP TRIGGER IF EXISTS items_history_trigger ON items;
DROP FUNCTION IF EXISTS log_item_changes();