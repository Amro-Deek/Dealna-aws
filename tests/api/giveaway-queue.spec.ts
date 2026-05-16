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
    const r1 = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: testEmail },
    });
    if (r1.status() !== 204) console.error('req-activation:', await r1.text());
    expect(r1.status()).toBe(204);

    await new Promise(ok => setTimeout(ok, 800));

    // 2. Read activation token from DB
    const activationToken = await dbHelper.getStudentActivationToken(testEmail);
    expect(activationToken).not.toBeNull();

    // 3. Activate
    const r2 = await request.get(`/api/v1/auth/student/activate?token=${activationToken}`);
    if (r2.status() !== 204) console.error('activate:', await r2.text());
    expect(r2.status()).toBe(204);

    // 4. Complete registration
    const r3 = await request.post('/api/v1/auth/student/complete', {
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
    
    // Seed a dummy owner to avoid "owner cannot join own queue" error
    const ownerRes = await dbHelper.getPool().query(`
      INSERT INTO "User" (email, auth_provider, status, role) 
      VALUES ('dummy_owner_' || $1 || '@birzeit.edu', 'KEYCLOAK', 'ACTIVE', 'STUDENT') 
      RETURNING user_id`, [ts]);
    const dummyOwnerId = ownerRes.rows[0].user_id;

    const categoryId = await dbHelper.seedTestCategory('Test Category');
    itemID = await dbHelper.seedTestItem(dummyOwnerId, categoryId, 'Test Giveaway Item', 0);

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
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/join`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Join failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('EntryID');
    createdEntryId = body.EntryID;
  });

  test('GET /giveaway/queue/:myItemId/position/:entryId should return position', async ({ request }) => {
    test.skip(!createdEntryId, 'Entry not created');
    const res = await request.get(`/api/v1/giveaway/queue/${itemID}/position/${createdEntryId}`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Position failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('position');
  });

  test('POST /giveaway/queue/:myItemId/leave should leave queue', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/leave`, {
      headers: { Authorization: `Bearer ${token}` },
    });
    expect(res.status(), `Leave failed: ${await res.text()}`).toBe(200);
  });
});

test.describe.serial('Giveaway Queue State Transitions API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const ownerEmail = `owner_${ts}@student.birzeit.edu`;
  const receiverEmail = `receiver_${ts}@student.birzeit.edu`;
  const testPassword = 'StrongPassword123!';

  let ownerToken: string;
  let receiverToken: string;
  let itemID: string;
  let entryID: string;

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // Setup Owner
    await request.post('/api/v1/auth/student/request-activation', { data: { email: ownerEmail } });
    await new Promise(ok => setTimeout(ok, 800));
    let t = await dbHelper.getStudentActivationToken(ownerEmail);
    await request.get(`/api/v1/auth/student/activate?token=${t}`);
    await request.post('/api/v1/auth/student/complete', {
      data: { email: ownerEmail, displayName: 'Owner', password: testPassword, major: 'CS', academicYear: 2 },
    });

    // Setup Receiver
    await request.post('/api/v1/auth/student/request-activation', { data: { email: receiverEmail } });
    await new Promise(ok => setTimeout(ok, 800));
    t = await dbHelper.getStudentActivationToken(receiverEmail);
    await request.get(`/api/v1/auth/student/activate?token=${t}`);
    await request.post('/api/v1/auth/student/complete', {
      data: { email: receiverEmail, displayName: 'Receiver', password: testPassword, major: 'CS', academicYear: 2 },
    });

    await new Promise(ok => setTimeout(ok, 800));

    ownerToken = await getStudentToken(ownerEmail, testPassword);
    receiverToken = await getStudentToken(receiverEmail, testPassword);

    const ownerId = await dbHelper.getUserIdByEmail(ownerEmail);
    if (!ownerId) throw new Error('Owner ID not found after registration');
    const categoryId = await dbHelper.seedTestCategory('Test Category');
    itemID = await dbHelper.seedTestItem(ownerId, categoryId, 'Transition Test Item', 0);

    await new Promise(ok => setTimeout(ok, 3000));
  });

  test.afterAll(async () => {
    await dbHelper.cleanupTestUser(ownerEmail);
    await dbHelper.cleanupTestUser(receiverEmail);
    try { await kcHelper.deleteUserByEmail(ownerEmail); } catch {}
    try { await kcHelper.deleteUserByEmail(receiverEmail); } catch {}
    await dbHelper.close();
  });

  test('Receiver joins queue', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/join`, {
      headers: { Authorization: `Bearer ${receiverToken}` },
    });
    expect(res.status(), `Join failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    entryID = body.EntryID;
    
    // First in queue is immediately promoted
    expect(body.EntryStatus).toBe('RESERVED');
  });

  test('Owner accepts turn', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entryID}/accept`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `Owner accept failed: ${await res.text()}`).toBe(200);
  });

  test('Owner initiates handoff', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entryID}/handoff`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `Owner handoff failed: ${await res.text()}`).toBe(200);
  });

  test('Receiver confirms handoff', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entryID}/complete`, {
      headers: { Authorization: `Bearer ${receiverToken}` },
    });
    expect(res.status(), `Receiver confirm failed: ${await res.text()}`).toBe(200);
  });
});

test.describe.serial('Giveaway Queue Negative and Edge Cases', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const ownerEmail = `owner_edge_${ts}@student.birzeit.edu`;
  const receiver1Email = `receiver1_edge_${ts}@student.birzeit.edu`;
  const receiver2Email = `receiver2_edge_${ts}@student.birzeit.edu`;
  const testPassword = 'StrongPassword123!';

  let ownerToken: string;
  let receiver1Token: string;
  let receiver2Token: string;
  let itemID: string;
  let entry1ID: string;
  let entry2ID: string;

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // Create 3 users
    for (const email of [ownerEmail, receiver1Email, receiver2Email]) {
      await request.post('/api/v1/auth/student/request-activation', { data: { email } });
      await new Promise(ok => setTimeout(ok, 800));
      const t = await dbHelper.getStudentActivationToken(email);
      await request.get(`/api/v1/auth/student/activate?token=${t}`);
      await request.post('/api/v1/auth/student/complete', {
        data: { email, displayName: email.split('_')[0], password: testPassword, major: 'CS', academicYear: 2 },
      });
      await new Promise(ok => setTimeout(ok, 800));
    }

    ownerToken = await getStudentToken(ownerEmail, testPassword);
    receiver1Token = await getStudentToken(receiver1Email, testPassword);
    receiver2Token = await getStudentToken(receiver2Email, testPassword);

    const ownerId = await dbHelper.getUserIdByEmail(ownerEmail);
    if (!ownerId) throw new Error('Owner ID not found');
    const categoryId = await dbHelper.seedTestCategory('Test Category');
    itemID = await dbHelper.seedTestItem(ownerId, categoryId, 'Edge Case Test Item', 0);

    await new Promise(ok => setTimeout(ok, 3000));
  });

  test.afterAll(async () => {
    for (const email of [ownerEmail, receiver1Email, receiver2Email]) {
      await dbHelper.cleanupTestUser(email);
      try { await kcHelper.deleteUserByEmail(email); } catch {}
    }
    await dbHelper.close();
  });

  test('Receiver 1 and 2 join the queue', async ({ request }) => {
    // Receiver 1 joins
    let res = await request.post(`/api/v1/giveaway/queue/${itemID}/join`, {
      headers: { Authorization: `Bearer ${receiver1Token}` },
    });
    expect(res.status(), `Join failed: ${await res.text()}`).toBe(200);
    let body = await res.json();
    entry1ID = body.EntryID;
    expect(body.EntryStatus).toBe('RESERVED');

    // Receiver 2 joins
    res = await request.post(`/api/v1/giveaway/queue/${itemID}/join`, {
      headers: { Authorization: `Bearer ${receiver2Token}` },
    });
    expect(res.status(), `Join failed: ${await res.text()}`).toBe(200);
    body = await res.json();
    entry2ID = body.EntryID;
    expect(body.EntryStatus).toBe('WAITING');
  });

  test('Unauthorized access: Receiver tries to accept turn', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entry1ID}/accept`, {
      headers: { Authorization: `Bearer ${receiver1Token}` },
    });
    expect(res.status()).toBe(500);
    expect(await res.text()).toContain('unauthorized');
  });

  test('State machine violation: Owner tries to handoff before accept', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entry1ID}/handoff`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status()).toBe(500);
    expect(await res.text()).toContain('entry is not in CONFIRMED state');
  });

  test('Owner rejects turn of Receiver 1', async ({ request }) => {
    // Owner rejects Receiver 1
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entry1ID}/reject`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `Owner reject failed: ${await res.text()}`).toBe(200);
  });

  test('Owner can accept Receiver 2 now', async ({ request }) => {
    const res = await request.post(`/api/v1/giveaway/queue/${itemID}/entries/${entry2ID}/accept`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    // This will only work if Receiver 2 is successfully promoted to RESERVED
    expect(res.status(), `Owner accept on Receiver 2 failed: ${await res.text()}`).toBe(200);
  });
});
