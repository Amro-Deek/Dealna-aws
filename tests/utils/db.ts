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

  constructor() {
    this.pool = new Pool({
      connectionString: process.env.DATABASE_URL,
    });
  }

  /**
   * Fetch the activation token generated during the student registration request.
   */
  async getStudentActivationToken(email: string): Promise<string | null> {
    const res = await this.pool.query(
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
    await this.pool.query(
      `INSERT INTO student_pre_registration (email, token, expires_at) VALUES ($1, $2, $3)`,
      [email, token, expiresAt]
    );
  }

  /**
   * Seed a verified student pre-registration (clicked the link, ready to complete).
   */
  async seedVerifiedPreRegistration(email: string): Promise<void> {
    const token = crypto.randomUUID(); // Use dynamic UUID to avoid unique constraint violations
    const expiresAt = new Date(Date.now() + 24 * 60 * 60 * 1000);
    const verifiedAt = new Date();
    await this.pool.query(
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
    await this.pool.query(
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
      // 1. Delete items created by this user (attachments cascade via FK)
      await this.pool.query(
        `DELETE FROM public.item WHERE owner_id = (SELECT user_id FROM public."User" WHERE email = $1 LIMIT 1)`,
        [email]
      );

      // 2. Delete from student_pre_registration
      await this.pool.query('DELETE FROM student_pre_registration WHERE email = $1', [email]);
      
      // 3. Delete from User table (which cascades down due to FKs)
      await this.pool.query('DELETE FROM "User" WHERE email = $1', [email]);
      
      console.log(`🧹 Cleaned up DB for user: ${email}`);
    } catch (e) {
      console.error(`❌ Failed to clean up DB for ${email}:`, e);
    }
  }

  /**
   * Ensure that the test university exists in the database.
   */
  async ensureUniversityExists(name: string, domain: string): Promise<void> {
    const res = await this.pool.query('SELECT university_id FROM university WHERE domain = $1', [domain]);
    if (res.rows.length === 0) {
      await this.pool.query(
        'INSERT INTO university (name, domain, status) VALUES ($1, $2, $3)',
        [name, domain, 'ACTIVE']
      );
      console.log(`🏢 Seeded university: ${name} (${domain})`);
    }
  }

  /**
   * Close the database connection pool.
   */
  async close(): Promise<void> {
    await this.pool.end();
  }
}
