SET search_path TO public;

-- 1. Add updated_at column for TTL tracking and cooldown enforcement
ALTER TABLE queue_entry
ADD COLUMN IF NOT EXISTS updated_at
  TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL;

-- 2. Backfill: set updated_at = joined_at for existing rows
UPDATE queue_entry SET updated_at = joined_at WHERE updated_at IS NULL;

-- 3. Drop the old blanket uniqueness constraint (blocks rejoining after cancel/expiry)
ALTER TABLE queue_entry DROP CONSTRAINT IF EXISTS queue_entry_item_id_user_id_key;

-- 4. Partial unique index: enforces active-entry uniqueness + accelerates queries
CREATE UNIQUE INDEX IF NOT EXISTS idx_queue_entry_active_unique
ON queue_entry (item_id, user_id)
WHERE entry_status IN ('WAITING', 'RESERVED', 'CONFIRMED', 'HANDED_OFF');

-- 5. Expiry worker index: fast lookups for background TTL checks
CREATE INDEX IF NOT EXISTS idx_queue_entry_expiry
ON queue_entry (entry_status, updated_at)
WHERE entry_status IN ('RESERVED', 'CONFIRMED', 'HANDED_OFF');
