import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

async function getStudentToken(email: string, password: string): Promise<string> {
  const base = (process.env.KEYCLOAK_BASE_URL ?? '').replace(/\/$/, '');
  const realm = process.env.KEYCLOAK_REALM ?? 'Dealna';
  const clientId = process.env.KEYCLOAK_ADMIN_CLIENT_ID ?? 'dealna-backend';
  const clientSecret = process.env.KEYCLOAK_ADMIN_CLIENT_SECRET ?? '';

  const params = new URLSearchParams({
    client_id: clientId,
    client_secret: clientSecret,
    grant_type: 'password',
    username: email,
    password,
  });

  const res = await fetch(`${base}/realms/${realm}/protocol/openid-connect/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: params.toString(),
  });

  if (!res.ok) throw new Error(`Keycloak login failed: ${await res.text()}`);
  const data = await res.json();
  return data.access_token as string;
}

test.describe.serial('Giveaway Queue API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const testEmail = `queue_tester_${ts}@student.birzeit.edu`;
  const testDisplay = `Queue Tester ${ts}`;
  const testPassword = 'StrongPassword123!';

  let token: string;
  let itemID: string;
  let createdEntryId = '';

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // 1. Request activation
    const r1 = await request.post('auth/student/request-activation', {
      data: { email: testEmail },
    });
    if (r1.status() !== 204) console.error('req-activation:', await r1.text());
    expect(r1.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    // 2. Read activation token from DB
    const activationToken = await dbHelper.getStudentActivationToken(testEmail);
    expect(activationToken).not.toBeNull();

    // 3. Activate
    const r2 = await request.get(`auth/student/activate?token=${activationToken}`);
    if (r2.status() !== 204) console.error('activate:', await r2.text());
    expect(r2.status()).toBe(204);

    // 4. Complete registration
    const r3 = await request.post('auth/student/complete', {
      data: {
        email: testEmail,
        displayName: testDisplay,
        password: testPassword,
        major: 'Computer Science',
        academicYear: 2,
      },
    });
    if (r3.status() !== 204) console.error('complete:', await r3.text());
    expect(r3.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    // 5. Login
    token = await getStudentToken(testEmail, testPassword);
    expect(token).toBeTruthy();

    // 6. Seed test item
    const userId = await dbHelper.getUserIdByEmail(testEmail);
    if (!userId) throw new Error('User ID not found after registration');
    const categoryId = await dbHelper.seedTestCategory('Test Category');
    itemID = await dbHelper.seedTestItem(userId, categoryId, 'Test Giveaway Item', 0);

    // Ensure backend/DB/Keycloak are fully synced before starting tests
    await new Promise(ok => setTimeout(ok, 3000));

    console.log(`✅ Queue tester ready: ${testEmail}`);
  });

  test.afterAll(async () => {
    await dbHelper.cleanupTestUser(testEmail);
    try { await kcHelper.deleteUserByEmail(testEmail); } catch { /* Keycloak may be unreachable */ }
    await dbHelper.close();
  });

  test('POST /giveaway/queue/:myItemId/join should join queue', async ({ request }) => {
    const res = await request.post(`giveaway/queue/${itemID}/join`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Join failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('EntryID');
    createdEntryId = body.EntryID;
  });

  test('GET /giveaway/queue/:myItemId/position/:entryId should return position', async ({ request }) => {
    test.skip(!createdEntryId, 'Entry not created');
    const res = await request.get(`giveaway/queue/${itemID}/position/${createdEntryId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Position failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('position');
  });

  test('POST /giveaway/queue/:myItemId/leave should leave queue', async ({ request }) => {
    const res = await request.post(`giveaway/queue/${itemID}/leave`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Leave failed: ${await res.text()}`).toBe(200);
  });
});
