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

test.describe.serial('Notifications API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const testEmail = `notif_tester_${ts}@student.birzeit.edu`;
  const testDisplay = `Notif Tester ${ts}`;
  const testPassword = 'StrongPassword123!';

  let token: string;
  let firstNotificationId = '';

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    const r1 = await request.post('/auth/student/request-activation', {
      data: { email: testEmail },
    });
    if (r1.status() !== 204) console.error('req-activation:', await r1.text());
    expect(r1.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    const activationToken = await dbHelper.getStudentActivationToken(testEmail);
    expect(activationToken).not.toBeNull();

    const r2 = await request.get(`/auth/student/activate?token=${activationToken}`);
    if (r2.status() !== 204) console.error('activate:', await r2.text());
    expect(r2.status()).toBe(204);

    const r3 = await request.post('/auth/student/complete', {
      data: {
        email: testEmail,
        displayName: testDisplay,
        password: testPassword,
        major: 'Architecture',
        academicYear: 4,
      },
    });
    if (r3.status() !== 204) console.error('complete:', await r3.text());
    expect(r3.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    token = await getStudentToken(testEmail, testPassword);
    expect(token).toBeTruthy();

    console.log(`✅ Notif tester ready: ${testEmail}`);
  });

  test.afterAll(async () => {
    await dbHelper.cleanupTestUser(testEmail);
    try { await kcHelper.deleteUserByEmail(testEmail); } catch { /* Keycloak may be unreachable */ }
    await dbHelper.close();
  });

  test('GET /giveaway/notifications should list notifications', async ({ request }) => {
    const res = await request.get('/giveaway/notifications', {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `List notifications failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBeTruthy();
    if (body.length > 0) firstNotificationId = body[0].id;
  });

  test('POST /giveaway/notifications/:notifId/read should mark notification as read', async ({ request }) => {
    test.skip(!firstNotificationId, 'No notifications found to mark read');
    const res = await request.post(`/giveaway/notifications/${firstNotificationId}/read`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Mark read failed: ${await res.text()}`).toBe(200);
  });
});
