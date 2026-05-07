SET search_path TO public;

DROP INDEX IF EXISTS idx_queue_entry_expiry;
DROP INDEX IF EXISTS idx_queue_entry_active_unique;
ALTER TABLE queue_entry ADD CONSTRAINT queue_entry_item_id_user_id_key UNIQUE (item_id, user_id);
ALTER TABLE queue_entry DROP COLUMN IF EXISTS updated_at;
