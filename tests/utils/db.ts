import { Pool } from 'pg';
import * as dotenv from 'dotenv';
import * as path from 'path';
import * as crypto from 'crypto';

// Ensure .env.test is loaded if running this utility independently
dotenv.config({ path: path.resolve(__dirname, '../.env.test') });

/**
 * Database Helper
 * Handles establishing a temporary connection to PostgreSQL to retrieve test-critical data and clean up test data.
 */
export class DatabaseHelper {
  private pool: Pool;
  private closed = false;

  constructor() {
    this.pool = new Pool({
      connectionString: process.env.DATABASE_URL,
    });
  }

  /**
   * Safely get the pool, reconnecting if it was previously closed.
   * Public to allow tests to run custom queries.
   */
  getPool(): Pool {
    if (this.closed) {
      this.pool = new Pool({
        connectionString: process.env.DATABASE_URL,
      });
      this.closed = false;
    }
    return this.pool;
  }

  /**
   * Fetch the activation token generated during the student registration request.
   */
  async getStudentActivationToken(email: string): Promise<string | null> {
    const res = await this.getPool().query(
      'SELECT token FROM student_pre_registration WHERE email = $1 LIMIT 1',
      [email]
    );
    if (res.rows.length > 0) {
      return res.rows[0].token;
    }
    return null;
  }

	/**
	 * Seed a pending student pre-registration (requested activation, not verified).
	 */
	async seedPendingPreRegistration(email: string, token: string): Promise<void> {
	  const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000); // Expires tomorrow
	  await this.getPool().query(
		`INSERT INTO student_pre_registration (email, token, expires_at) VALUES ($1, $2, $3)`,
		[email, token, expiresAt]
	  );
	}

	async seedAdminUser(email: string, keycloakSub: string): Promise<void> {
		const res = await this.getPool().query(
			`INSERT INTO "User" (email, role, account_status, email_verified, university_id, keycloak_sub)
			VALUES ($1, 'ADMIN', 'ACTIVE', true, '00000000-0000-0000-0000-000000000000', $2)
			ON CONFLICT (email) DO UPDATE SET role = 'ADMIN', keycloak_sub = EXCLUDED.keycloak_sub
			RETURNING user_id`,
			[email, keycloakSub]
		);
		const userId = res.rows[0].user_id;

		await this.getPool().query(
			`INSERT INTO admin (user_id) VALUES ($1) ON CONFLICT (user_id) DO NOTHING`,
			[userId]
		);
	}

  /**
   * Seed a verified student pre-registration (clicked the link, ready to complete).
   */
  async seedVerifiedPreRegistration(email: string): Promise<void> {
    const token = crypto.randomUUID(); // Use dynamic UUID to avoid unique constraint violations
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
    const verifiedAt = new Date();
    await this.getPool().query(
      `INSERT INTO student_pre_registration (email, token, expires_at, verified_at) VALUES ($1, $2, $3, $4)`,
      [email, token, expiresAt, verifiedAt]
    );
  }

  /**
   * Seed a completed student pre-registration.
   */
  async seedCompletedPreRegistration(email: string): Promise<void> {
    const token = crypto.randomUUID(); // Use dynamic UUID
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
    const verifiedAt = new Date();
    const usedAt = new Date();
    await this.getPool().query(
      `INSERT INTO student_pre_registration (email, token, expires_at, verified_at, used_at) VALUES ($1, $2, $3, $4, $5)`,
      [email, token, expiresAt, verifiedAt, usedAt]
    );
  }

  /**
   * Delete test user data from the database to maintain test isolation.
   * This removes pre-registration and the main user account (cascading to profiles, students, etc.).
   */
  async cleanupTestUser(email: string): Promise<void> {
    try {
      const pool = this.getPool();
      // 1. Delete items created by this user (attachments cascade via FK)
      await pool.query(
        `DELETE FROM public.item WHERE owner_id = (SELECT user_id FROM public."User" WHERE email = $1 LIMIT 1)`,
        [email]
      );

      // 2. Delete from student_pre_registration
      await pool.query('DELETE FROM student_pre_registration WHERE email = $1', [email]);

      // 3. Delete from User table (which cascades down due to FKs)
      await pool.query('DELETE FROM "User" WHERE email = $1', [email]);

      console.log(`🧹 Cleaned up DB for user: ${email}`);
    } catch (e) {
      console.error(`❌ Failed to clean up DB for ${email}:`, e);
    }
  }

  /**
   * Seed a test category if it doesn't exist.
   */
  async seedTestCategory(name: string): Promise<string> {
    const pool = this.getPool();
    const res = await pool.query('SELECT category_id FROM category WHERE name = $1', [name]);
    if (res.rows.length === 0) {
      const insertRes = await pool.query(
        'INSERT INTO category (name, description) VALUES ($1, $2) RETURNING category_id',
        [name, `Test category for ${name}`]
      );
      return insertRes.rows[0].category_id.toString();
    }
    return res.rows[0].category_id;
  }

  /**
   * Seed a test item for a given owner and category.
   */
  async seedTestItem(ownerId: string, categoryId: string, title: string, price: number): Promise<string> {
    const pool = this.getPool();
    const insertRes = await pool.query(
      `INSERT INTO item (owner_id, category_id, title, description, price, item_status) 
       VALUES ($1, $2, $3, $4, $5, $6) RETURNING item_id`,
      [ownerId, categoryId, title, 'Test item description', price, 'AVAILABLE']
    );
    return insertRes.rows[0].item_id.toString();
  }

  /**
   * Seed a dummy user with a profile (needed so GetItemDetails JOIN on profile succeeds).
   * Returns the user_id of the created user.
   */
  async seedDummyUser(emailPrefix: string, displayName: string): Promise<string> {
    const pool = this.getPool();
    // Step 1: insert User
    const userRes = await pool.query(
      `INSERT INTO "User" (email, role, account_status, university_id) 
       SELECT $1, 'STUDENT', 'ACTIVE', university_id 
       FROM university WHERE domain = 'birzeit.edu'
       RETURNING user_id`,
      [emailPrefix]
    );
    const userId = userRes.rows[0].user_id;

    // Step 2: insert Profile for that user
    await pool.query(
      `INSERT INTO profile (user_id, display_name) VALUES ($1, $2)`,
      [userId, displayName]
    );

    return userId;
  }

  /**
   * Get user_id by email.
   */
  async getUserIdByEmail(email: string): Promise<string | null> {
    const res = await this.getPool().query('SELECT user_id FROM "User" WHERE email = $1', [email]);
    if (res.rows.length > 0) return res.rows[0].user_id;
    return null;
  }

  /**
   * Ensure that the test university exists in the database.
   */
  async ensureUniversityExists(name: string, domain: string): Promise<void> {
    const pool = this.getPool();
    const res = await pool.query('SELECT university_id FROM university WHERE domain = $1', [domain]);
    if (res.rows.length === 0) {
      await pool.query(
        'INSERT INTO university (name, domain, status) VALUES ($1, $2, $3) ON CONFLICT (domain) DO NOTHING',
        [name, domain, 'ACTIVE']
      );
      console.log(`🏢 Ensured university exists: ${name} (${domain})`);
    }
  }

  /**
   * Close the database connection pool.
   */
  async close(): Promise<void> {
    if (!this.closed) {
      this.closed = true;
      await this.pool.end();
    }
  }
}
