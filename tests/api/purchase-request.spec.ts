import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

// ─── Shared helper ────────────────────────────────────────────────────────────
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

async function registerStudent(
  request: any,
  dbHelper: DatabaseHelper,
  email: string,
  displayName: string,
  password: string
): Promise<void> {
  const r1 = await request.post('/api/v1/auth/student/request-activation', { data: { email } });
  if (r1.status() !== 204) throw new Error(`request-activation failed: ${await r1.text()}`);

  await new Promise(ok => setTimeout(ok, 800));

  const token = await dbHelper.getStudentActivationToken(email);
  if (!token) throw new Error('Activation token not found');

  const r2 = await request.get(`/api/v1/auth/student/activate?token=${token}`);
  if (r2.status() !== 204) throw new Error(`activate failed: ${await r2.text()}`);

  const r3 = await request.post('/api/v1/auth/student/complete', {
    data: { email, displayName, password, major: 'Computer Science', academicYear: 2 },
  });
  if (r3.status() !== 204) throw new Error(`complete failed: ${await r3.text()}`);

  await new Promise(ok => setTimeout(ok, 800));
}

// ─── Suite 1: Happy Path ───────────────────────────────────────────────────────
test.describe.serial('Purchase Requests — Happy Path', () => {
  const db = new DatabaseHelper();
  const kc = new KeycloakHelper();
  const ts = Date.now();
  const password = 'StrongPassword123!';

  const buyerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const ownerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;

  let buyerToken: string;
  let ownerToken: string;
  let itemID: string;
  let requestID: string;

  test.beforeAll(async ({ request }) => {
    await db.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // Register buyer (goes through Keycloak)
    await registerStudent(request, db, buyerEmail, `Buyer ${ts}`, password);

    // Owner is a DB-seeded dummy (no Keycloak account needed for owning an item)
    const ownerId = await db.seedDummyUser(
      `dummy_purchase_owner_${ts}@birzeit.edu`,
      `Owner ${ts}`
    );

    // Seed item owned by dummy owner
    const categoryId = await db.seedTestCategory('Electronics & Tech');
    itemID = await db.seedTestItem(ownerId, categoryId, `Sale Item ${ts}`, 150);

    // Get buyer token
    buyerToken = await getStudentToken(buyerEmail, password);
    expect(buyerToken).toBeTruthy();

    // Register owner through Keycloak so they can also call APIs
    await registerStudent(request, db, ownerEmail, `Owner KC ${ts}`, password);
    ownerToken = await getStudentToken(ownerEmail, password);

    // Re-seed item under the KC owner so accept/reject works
    const ownerKcId = await db.getUserIdByEmail(ownerEmail);
    if (!ownerKcId) throw new Error('KC owner not found');
    itemID = await db.seedTestItem(ownerKcId, categoryId, `Sale Item KC ${ts}`, 150);

    console.log(`✅ Purchase happy path ready — buyer: ${buyerEmail}, owner: ${ownerEmail}`);
  });

  test.afterAll(async () => {
    await db.cleanupTestUser(buyerEmail);
    await db.cleanupTestUser(ownerEmail);
    try { await kc.deleteUserByEmail(buyerEmail); } catch { }
    try { await kc.deleteUserByEmail(ownerEmail); } catch { }
    await db.close();
  });

  test('POST /purchases/items/:id/request — buyer can request', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Create request failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('RequestID');
    requestID = body.RequestID;
  });

  test('GET /purchases/items/:id/requests — owner lists requests', async ({ request }) => {
    const res = await request.get(`/api/v1/purchases/items/${itemID}/requests`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `List requests failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBeTruthy();
    expect(body.length).toBeGreaterThan(0);
  });

  test('GET /purchases/me — buyer sees their own requests', async ({ request }) => {
    const res = await request.get(`/api/v1/purchases/me`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Get my requests failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBeTruthy();
  });

  test('POST /purchases/items/:id/requests/:reqId/accept — owner accepts', async ({ request }) => {
    test.skip(!requestID, 'Request was not created');
    const res = await request.post(`/api/v1/purchases/items/${itemID}/requests/${requestID}/accept`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `Accept failed: ${await res.text()}`).toBe(200);
  });
});

// ─── Suite 2: Reject Flow ──────────────────────────────────────────────────────
test.describe.serial('Purchase Requests — Reject Flow', () => {
  const db = new DatabaseHelper();
  const kc = new KeycloakHelper();
  const ts = Date.now() + 1;
  const password = 'StrongPassword123!';

  const buyerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const ownerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;

  let buyerToken: string;
  let ownerToken: string;
  let itemID: string;
  let requestID: string;

  test.beforeAll(async ({ request }) => {
    await db.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    await registerStudent(request, db, buyerEmail, `Buyer Rej ${ts}`, password);
    await registerStudent(request, db, ownerEmail, `Owner Rej ${ts}`, password);

    buyerToken = await getStudentToken(buyerEmail, password);
    ownerToken = await getStudentToken(ownerEmail, password);

    const ownerId = await db.getUserIdByEmail(ownerEmail);
    if (!ownerId) throw new Error('Owner not found');
    const categoryId = await db.seedTestCategory('Electronics & Tech');
    itemID = await db.seedTestItem(ownerId, categoryId, `Reject Item ${ts}`, 200);
  });

  test.afterAll(async () => {
    await db.cleanupTestUser(buyerEmail);
    await db.cleanupTestUser(ownerEmail);
    try { await kc.deleteUserByEmail(buyerEmail); } catch { }
    try { await kc.deleteUserByEmail(ownerEmail); } catch { }
    await db.close();
  });

  test('Buyer creates request', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Create failed: ${await res.text()}`).toBe(200);
    const body = await res.json();
    requestID = body.RequestID;
  });

  test('Owner rejects request', async ({ request }) => {
    test.skip(!requestID, 'Request not created');
    const res = await request.post(`/api/v1/purchases/items/${itemID}/requests/${requestID}/reject`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status(), `Reject failed: ${await res.text()}`).toBe(200);
  });
});

// ─── Suite 3: Cancel Flow ──────────────────────────────────────────────────────
test.describe.serial('Purchase Requests — Cancel Flow (Buyer)', () => {
  const db = new DatabaseHelper();
  const kc = new KeycloakHelper();
  const ts = Date.now() + 2;
  const password = 'StrongPassword123!';

  const buyerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const ownerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;

  let buyerToken: string;
  let ownerToken: string;
  let itemID: string;
  let requestID: string;

  test.beforeAll(async ({ request }) => {
    await db.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    await registerStudent(request, db, buyerEmail, `Buyer Cancel ${ts}`, password);
    await registerStudent(request, db, ownerEmail, `Owner Cancel ${ts}`, password);

    buyerToken = await getStudentToken(buyerEmail, password);
    ownerToken = await getStudentToken(ownerEmail, password);

    const ownerId = await db.getUserIdByEmail(ownerEmail);
    if (!ownerId) throw new Error('Owner not found');
    const categoryId = await db.seedTestCategory('Electronics & Tech');
    itemID = await db.seedTestItem(ownerId, categoryId, `Cancel Item ${ts}`, 300);
  });

  test.afterAll(async () => {
    await db.cleanupTestUser(buyerEmail);
    await db.cleanupTestUser(ownerEmail);
    try { await kc.deleteUserByEmail(buyerEmail); } catch { }
    try { await kc.deleteUserByEmail(ownerEmail); } catch { }
    await db.close();
  });

  test('Buyer creates request', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Create failed: ${await res.text()}`).toBe(200);
    requestID = (await res.json()).RequestID;
  });

  test('Buyer cancels their own request', async ({ request }) => {
    test.skip(!requestID, 'Request not created');
    const res = await request.post(`/api/v1/purchases/items/${itemID}/requests/${requestID}/cancel`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Cancel failed: ${await res.text()}`).toBe(200);
  });
});

// ─── Suite 4: Security & Edge Cases ───────────────────────────────────────────
test.describe.serial('Purchase Requests — Security & Edge Cases', () => {
  const db = new DatabaseHelper();
  const kc = new KeycloakHelper();
  const ts = Date.now() + 3;
  const password = 'StrongPassword123!';

  const ownerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const buyerEmail = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;

  let ownerToken: string;
  let buyerToken: string;
  let itemID: string;
  let requestID: string;

  test.beforeAll(async ({ request }) => {
    await db.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    await registerStudent(request, db, ownerEmail, `Owner Sec ${ts}`, password);
    await registerStudent(request, db, buyerEmail, `Buyer Sec ${ts}`, password);

    ownerToken = await getStudentToken(ownerEmail, password);
    buyerToken = await getStudentToken(buyerEmail, password);

    const ownerId = await db.getUserIdByEmail(ownerEmail);
    if (!ownerId) throw new Error('Owner not found');
    const categoryId = await db.seedTestCategory('Electronics & Tech');
    itemID = await db.seedTestItem(ownerId, categoryId, `Security Test Item ${ts}`, 99);
  });

  test.afterAll(async () => {
    await db.cleanupTestUser(ownerEmail);
    await db.cleanupTestUser(buyerEmail);
    try { await kc.deleteUserByEmail(ownerEmail); } catch { }
    try { await kc.deleteUserByEmail(buyerEmail); } catch { }
    await db.close();
  });

  test('❌ Owner cannot buy their own item', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${ownerToken}` },
    });
    expect(res.status()).toBe(500);
    expect(await res.text()).toContain('cannot purchase your own item');
  });

  test('✅ Buyer can send a request', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status(), `Create failed: ${await res.text()}`).toBe(200);
    requestID = (await res.json()).RequestID;
  });

  test('❌ Buyer cannot send duplicate request', async ({ request }) => {
    test.skip(!requestID, 'First request not created');
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status()).toBe(500);
    expect(await res.text()).toContain('already have an active purchase request');
  });

  test('❌ Buyer cannot accept their own request (only owner can)', async ({ request }) => {
    test.skip(!requestID, 'Request not created');
    const res = await request.post(`/api/v1/purchases/items/${itemID}/requests/${requestID}/accept`, {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });
    expect(res.status()).toBe(403);
  });

  test('❌ Unauthenticated request is rejected', async ({ request }) => {
    const res = await request.post(`/api/v1/purchases/items/${itemID}/request`);
    expect(res.status()).toBe(401);
  });
});
