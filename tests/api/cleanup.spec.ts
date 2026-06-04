import { test } from '@playwright/test';
import * as dotenv from 'dotenv';
import * as path from 'path';

// Force load backend env
dotenv.config({ path: path.resolve(__dirname, '../../backend/.env'), override: true });

import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

test('Wipe out student 1222375 wildcard', async () => {
  const db = new DatabaseHelper();
  
  // Find all users matching %1222375%
  const res = await db.getPool().query('SELECT email FROM "User" WHERE email LIKE $1', ['%1222375%']);
  console.log('Found users:', res.rows);
  
  for (const row of res.rows) {
    const email = row.email;
    console.log('Cleaning up user:', email);
    await db.cleanupTestUser(email);
    const kc = new KeycloakHelper();
    await kc.deleteUserByEmail(email);
  }

  // Find all pre-registrations matching %1222375%
  const preRes = await db.getPool().query('SELECT email FROM student_pre_registration WHERE email LIKE $1', ['%1222375%']);
  console.log('Found pre-registrations:', preRes.rows);
  
  for (const row of preRes.rows) {
    const email = row.email;
    await db.getPool().query('DELETE FROM student_pre_registration WHERE email = $1', [email]);
    console.log('Cleaned up lingering pre-registration for ' + email);
  }

  await db.close();
});
