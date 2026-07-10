import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

// ─────────────────────────────────────────────────────────
// Helper: get a student token via Keycloak password grant
// ─────────────────────────────────────────────────────────
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

test.describe.serial('Rating System API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  
  // Buyer
  const buyerEmail    = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const testPassword  = 'StrongPassword123!';
  const buyerDisplay  = `Buyer Rate Tester ${ts}`;
  let buyerToken: string;
  let buyerUserId: string;

  // Seller
  const sellerEmail   = `${Math.floor(1000000 + Math.random() * 9000000)}@student.birzeit.edu`;
  const sellerDisplay = `Seller Rate Tester ${ts}`;
  let sellerUserId: string;

  let itemId: string;
  let transactionId: string;

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // ── 1. Create the Seller Dummy directly in DB to save time ──
    sellerUserId = await dbHelper.seedDummyUser(sellerEmail, sellerDisplay);
    
    // ── 2. Create an Item for the Seller ──
    const catId = await dbHelper.seedTestCategory('Test Rate Category');
    itemId = await dbHelper.seedTestItem(sellerUserId, catId, 'Item to be Rated', 50.0);

    // ── 3. Register the Buyer properly so we can get a JWT ──
    await request.post('/api/v1/auth/student/request-activation', { data: { email: buyerEmail } });
    await new Promise(ok => setTimeout(ok, 500));
    const token = await dbHelper.getStudentActivationToken(buyerEmail);
    await request.get(`/api/v1/auth/student/activate?token=${token}`);
    await request.post('/api/v1/auth/student/complete', {
      data: {
        email: buyerEmail,
        displayName: buyerDisplay,
        password: testPassword,
        major: 'Computer Science',
        academicYear: 3,
      },
    });
    await new Promise(ok => setTimeout(ok, 500));

    buyerToken = await getStudentToken(buyerEmail, testPassword);
    const buyerIdRaw = await dbHelper.getUserIdByEmail(buyerEmail);
    if (!buyerIdRaw) throw new Error('Buyer ID not found');
    buyerUserId = buyerIdRaw;

    // ── 4. Seed a Completed Transaction that is 15 days old ──
    const pool = dbHelper.getPool();
    const txRes = await pool.query(
      `INSERT INTO transaction (item_id, buyer_id, seller_id, transaction_status, completed_at)
       VALUES ($1, $2, $3, 'COMPLETED', CURRENT_TIMESTAMP - INTERVAL '15 days')
       RETURNING transaction_id`,
      [itemId, buyerUserId, sellerUserId]
    );
    transactionId = txRes.rows[0].transaction_id;

    console.log(`✅ Test Environment Ready. Transaction ID: ${transactionId}`);
  });

  // ── Test 1: Fetch Pending Ratings (Should trigger mandatory lock) ──
  test('GET /users/me/pending-ratings — returns the 15-day old transaction', async ({ request }) => {
    const res = await request.get('/api/v1/users/me/pending-ratings', {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    
    // Should be an array with 1 item
    expect(Array.isArray(body)).toBe(true);
    expect(body.length).toBe(1);
    
    const pending = body[0];
    expect(pending.transaction_id).toBe(transactionId);
    expect(pending.seller_name).toBe(sellerDisplay);
    expect(pending.days_since_completion).toBeGreaterThanOrEqual(15);
  });

  // ── Test 2: Submit a 5-Star Rating ──
  test('POST /transactions/{id}/rate — successfully submits a rating', async ({ request }) => {
    const res = await request.post(`/api/v1/transactions/${transactionId}/rate`, {
      headers: { 
        Authorization: `Bearer ${buyerToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        stars: 5,
        comment: 'Great seller, fast communication!',
      },
    });

    if (res.status() !== 201) console.error('rate fail:', await res.text());
    expect(res.status()).toBe(201);

    const body = await res.json();
    expect(body.rating_id).toBeTruthy();
    expect(body.stars).toBe(5);
  });

  // ── Test 3: Fetch Pending Ratings Again (Should be empty) ──
  test('GET /users/me/pending-ratings — returns empty array after rating', async ({ request }) => {
    const res = await request.get('/api/v1/users/me/pending-ratings', {
      headers: { Authorization: `Bearer ${buyerToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    
    // Should now be empty, releasing the lock on the UI
    expect(Array.isArray(body)).toBe(true);
    expect(body.length).toBe(0);
  });

  // ── Test 4: Prevent Duplicate Ratings ──
  test('POST /transactions/{id}/rate — prevents double rating', async ({ request }) => {
    const res = await request.post(`/api/v1/transactions/${transactionId}/rate`, {
      headers: { 
        Authorization: `Bearer ${buyerToken}`,
        'Content-Type': 'application/json',
      },
      data: { stars: 4, comment: 'Oops I want to change it' },
    });

    // We expect a Bad Request or Conflict since they already rated
    expect(res.status()).toBe(400); 
  });

  // ── Teardown ──
  test.afterAll(async () => {
    const pool = dbHelper.getPool();
    // Delete Rating first (FK constraint)
    await pool.query(`DELETE FROM rating WHERE transaction_id = $1`, [transactionId]);
    // Delete Transaction (FK constraint)
    await pool.query(`DELETE FROM transaction WHERE transaction_id = $1`, [transactionId]);

    // Clean up users
    await kcHelper.deleteUserByEmail(buyerEmail);
    await dbHelper.cleanupTestUser(buyerEmail);
    await dbHelper.cleanupTestUser(sellerEmail);
    await dbHelper.close();
  });
});
