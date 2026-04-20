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

test.describe.serial('Purchase Requests API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const testEmail = `purchase_tester_${ts}@student.birzeit.edu`;
  const testDisplay = `Purchase Tester ${ts}`;
  const testPassword = 'StrongPassword123!';

  let token: string;
  let itemID: string;

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    const r1 = await request.post('auth/student/request-activation', {
      data: { email: testEmail },
    });
    if (r1.status() !== 204) console.error('req-activation:', await r1.text());
    expect(r1.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    const activationToken = await dbHelper.getStudentActivationToken(testEmail);
    expect(activationToken).not.toBeNull();

    const r2 = await request.get(`auth/student/activate?token=${activationToken}`);
    if (r2.status() !== 204) console.error('activate:', await r2.text());
    expect(r2.status()).toBe(204);

    const r3 = await request.post('auth/student/complete', {
      data: {
        email: testEmail,
        displayName: testDisplay,
        password: testPassword,
        major: 'Computer Engineering',
        academicYear: 3,
      },
    });
    if (r3.status() !== 204) console.error('complete:', await r3.text());
    expect(r3.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    token = await getStudentToken(testEmail, testPassword);
    expect(token).toBeTruthy();

    const userId = await dbHelper.getUserIdByEmail(testEmail);
    if (!userId) throw new Error('User ID not found');
    const categoryId = await dbHelper.seedTestCategory('Electronics');
    itemID = await dbHelper.seedTestItem(userId, categoryId, 'Test Sale Item', 100);

    console.log(`✅ Purchase tester ready: ${testEmail}`);
  });

  test.afterAll(async () => {
    await dbHelper.cleanupTestUser(testEmail);
    try { await kcHelper.deleteUserByEmail(testEmail); } catch { /* Keycloak may be unreachable */ }
    await dbHelper.close();
  });

  test('POST /giveaway/purchase/:myItemId/request should create a purchase request', async ({ request }) => {
    const res = await request.post(`giveaway/purchase/${itemID}/request`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Purchase request failed: ${await res.text()}`).toBe(200);
  });

  test('GET /giveaway/purchase/:myItemId/requests should return all requests', async ({ request }) => {
    const res = await request.get(`giveaway/purchase/${itemID}/requests`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `List requests failed: ${await res.text()}`).toBe(200);
  });
});
