import { Pool } from 'pg';
import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({ path: path.resolve(__dirname, '.env.test') });

const pool = new Pool({ connectionString: process.env.DATABASE_URL });

async function run() {
  await pool.query(`
    CREATE TABLE IF NOT EXISTS public.provider_pre_registration (
        id uuid DEFAULT gen_random_uuid() NOT NULL,
        email character varying(255) NOT NULL,
        token uuid NOT NULL,
        expires_at timestamp without time zone NOT NULL,
        used_at timestamp without time zone,
        resend_count integer DEFAULT 0 NOT NULL,
        resend_window_start timestamp without time zone,
        created_at timestamp without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
        verified_at timestamp without time zone,
        UNIQUE(email),
        UNIQUE(token)
    );
    ALTER TABLE public.provider_pre_registration OWNER TO dealna_user;
  `);
  console.log('Migration successful');
  process.exit(0);
}

run().catch(console.error);
