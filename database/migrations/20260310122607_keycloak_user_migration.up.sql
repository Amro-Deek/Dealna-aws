BEGIN;

-- 1️⃣ add keycloak_sub to link Keycloak user
ALTER TABLE public."User"
ADD COLUMN keycloak_sub UUID UNIQUE;

-- 2️⃣ remove local auth data (optional but recommended)
ALTER TABLE public."User"
DROP COLUMN IF EXISTS password_hash,
DROP COLUMN IF EXISTS failed_login_attempts,
DROP COLUMN IF EXISTS last_login_at;

-- 3️⃣ drop sessions table (Keycloak will manage sessions)
DROP TABLE IF EXISTS sessions;

COMMIT;