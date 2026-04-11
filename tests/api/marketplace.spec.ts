import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

// ─────────────────────────────────────────────────────────
// Test image — small JPEG from Designs
// ─────────────────────────────────────────────────────────
const TEST_IMAGE_PATH = 'C:\\Users\\fd\\Desktop\\Designs\\design.jpg';

// ─────────────────────────────────────────────────────────
// Helper: get a student token via Keycloak password grant
// ─────────────────────────────────────────────────────────
async function getStudentToken(email: string, password: string): Promise<string> {
  const base = (process.env.KEYCLOAK_BASE_URL ?? '').replace(/\/$/, '');
  const realm = process.env.KEYCLOAK_REALM ?? 'Dealna';
  // Use the confidential backend client for password grants in tests.
  // The public mobile client may not have Direct Access Grants enabled.
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

// ─────────────────────────────────────────────────────────
// Helper: PUT image file directly to S3 presigned URL
// ─────────────────────────────────────────────────────────
async function uploadToS3(presignedUrl: string, imagePath: string): Promise<void> {
  const buffer = fs.readFileSync(imagePath);
  const res = await fetch(presignedUrl, {
    method: 'PUT',
    headers: { 'Content-Type': 'image/jpeg' },
    body: buffer,
  });
  if (!res.ok) throw new Error(`S3 upload failed (${res.status}): ${await res.text()}`);
}

// ─────────────────────────────────────────────────────────
// Suite
// ─────────────────────────────────────────────────────────
test.describe.serial('Marketplace Feed & Item Posting API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts           = Date.now();
  const testEmail    = `test_market_${ts}@student.birzeit.edu`;
  const testPassword = 'StrongPassword123!';
  const testDisplay  = `Market Tester ${ts}`;

  let accessToken: string;
  let objectKey: string;
  let createdItemId: string;

  // ── beforeAll: register a throwaway student ──
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

    // 3. Activate email link
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

    // 5. Get JWT via Keycloak
    accessToken = await getStudentToken(testEmail, testPassword);
    expect(accessToken).toBeTruthy();
    console.log(`✅ Test student ready: ${testEmail}`);
  });

  // ── 1. Categories — public ──
  test('GET /categories — 200 public, returns seeded categories', async ({ request }) => {
    const res = await request.get('/api/v1/categories');

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBe(true);
    expect(body.length).toBeGreaterThan(0);

    const names: string[] = body.map((c: any) => c.name);
    console.log('📂 Categories:', names.join(', '));
    expect(names).toContain('Academic Materials');
    expect(names).toContain('Electronics & Tech');
    expect(names).toContain('Tutoring & Academic Services');
  });

  // ── 2. Feed without token ──
  test('GET /items/feed — 401 when no bearer token', async ({ request }) => {
    const res = await request.get('/api/v1/items/feed');
    expect(res.status()).toBe(401);
    const body = await res.json();
    expect(body.code).toBe('UNAUTHORIZED');
  });

  // ── 3. Presigned upload URL + real S3 upload ──
  test('POST /items/picture/upload-url — returns valid presigned URL, upload succeeds', async ({ request }) => {
    const res = await request.post('/api/v1/items/picture/upload-url?content_type=image/jpeg', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.upload_url).toBeTruthy();
    expect(body.object_key).toMatch(/^items\/.+\.png$/);
    objectKey = body.object_key;
    console.log(`🔑 Object key: ${objectKey}`);

    await uploadToS3(body.upload_url, TEST_IMAGE_PATH);
    console.log('✅ Image uploaded to S3');
  });

  // ── 4. Create listing ──
  test('POST /items — 201 creates listing with image', async ({ request }) => {
    expect(objectKey).toBeTruthy();

    const res = await request.post('/api/v1/items', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        title: 'Calculus Textbook 3rd Edition',
        description: 'Good condition, used one semester.',
        price: 30.0,
        pickup_location: 'BZU Main Library',
        object_keys: [objectKey],
      },
    });

    if (res.status() !== 201) console.error('create item:', await res.text());
    expect(res.status()).toBe(201);

    const body = await res.json();
    expect(body.id).toBeTruthy();
    expect(body.title).toBe('Calculus Textbook 3rd Edition');
    expect(body.status).toBe('AVAILABLE');

    createdItemId = body.id;
    console.log(`📦 Item created: ${createdItemId}`);
  });

  // ── 5. Validation: title too short ──
  test('POST /items — 400 when title is fewer than 5 chars', async ({ request }) => {
    const res = await request.post('/api/v1/items', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        title: 'Hi',
        price: 10.0,
        object_keys: [objectKey],
      },
    });

    expect(res.status()).toBe(400);
    const body = await res.json();
    expect(body.code).toBe('VALIDATION_FAILED');
  });

  // ── 6. Validation: negative price ──
  test('POST /items — 400 when price is negative', async ({ request }) => {
    const res = await request.post('/api/v1/items', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        title: 'Valid Title Long Enough',
        price: -5.0,
        object_keys: [objectKey],
      },
    });

    expect(res.status()).toBe(400);
    const body = await res.json();
    expect(body.code).toBe('VALIDATION_FAILED');
  });

  // ── 7. Validation: no images ──
  test('POST /items — 400 when object_keys is empty', async ({ request }) => {
    const res = await request.post('/api/v1/items', {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: {
        title: 'Valid Title Long Enough',
        price: 10.0,
        object_keys: [],
      },
    });

    expect(res.status()).toBe(400);
  });

  // ── 8. Feed — university scoped ──
  test('GET /items/feed — 200 contains newly created item', async ({ request }) => {
    const res = await request.get('/api/v1/items/feed?limit=20', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBe(true);

    const found = body.find((i: any) => i.id === createdItemId);
    expect(found).toBeTruthy();
    expect(found.status).toBe('AVAILABLE');
    console.log(`📰 Feed: ${body.length} item(s). Our item visible ✅`);
  });

  // ── 9. Feed — search ──
  test('GET /items/feed?search=Calculus — only matching items', async ({ request }) => {
    const res = await request.get('/api/v1/items/feed?search=Calculus', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBe(true);
    for (const item of body) {
      expect(item.title.toLowerCase()).toContain('calculus');
    }
  });

  // ── 10. My listings (storefront) ──
  test('GET /items/my — 200 returns only caller\'s items', async ({ request }) => {
    const res = await request.get('/api/v1/items/my', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBe(true);
    expect(body.length).toBeGreaterThan(0);

    const found = body.find((i: any) => i.id === createdItemId);
    expect(found).toBeTruthy();
    console.log(`🏪 Storefront: ${body.length} item(s)`);
  });

  // ── 11. Item detail with attachments ──
  test('GET /items/{id} — 200 returns item with attachment', async ({ request }) => {
    const res = await request.get(`/api/v1/items/${createdItemId}`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.id).toBe(createdItemId);
    expect(Array.isArray(body.attachments)).toBe(true);
    expect(body.attachments.length).toBeGreaterThan(0);
    expect(body.attachments[0].file_path).toBe(objectKey);
    console.log(`🖼️  Attachments: ${body.attachments.map((a: any) => a.file_path).join(', ')}`);
  });

  // ── 12. Status → RESERVED ──
  test('PATCH /items/{id}/status — 200 transitions to RESERVED', async ({ request }) => {
    const res = await request.patch(`/api/v1/items/${createdItemId}/status`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: { status: 'RESERVED' },
    });

    expect(res.status()).toBe(200);
    expect((await res.json()).message).toBe('status updated');
    console.log('✅ Status → RESERVED');
  });

  // ── 13. Status → SOLD ──
  test('PATCH /items/{id}/status — 200 transitions to SOLD', async ({ request }) => {
    const res = await request.patch(`/api/v1/items/${createdItemId}/status`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: { status: 'SOLD' },
    });

    expect(res.status()).toBe(200);
    console.log('✅ Status → SOLD');
  });

  // ── 14. Invalid status ──
  test('PATCH /items/{id}/status — 400 for invalid status', async ({ request }) => {
    const res = await request.patch(`/api/v1/items/${createdItemId}/status`, {
      headers: {
        Authorization: `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      data: { status: 'GARBAGE' },
    });

    expect(res.status()).toBe(400);
    expect((await res.json()).code).toBe('VALIDATION_FAILED');
  });

  // ── 15. Soft delete ──
  test('DELETE /items/{id} — 200 soft-deletes the listing', async ({ request }) => {
    const res = await request.delete(`/api/v1/items/${createdItemId}`, {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    expect((await res.json()).message).toBe('item deleted');
    console.log('✅ Item soft-deleted');
  });

  // ── 16. Deleted item removed from feed ──
  test('GET /items/feed — deleted item is no longer visible', async ({ request }) => {
    const res = await request.get('/api/v1/items/feed?limit=50', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    const found = body.find((i: any) => i.id === createdItemId);
    expect(found).toBeUndefined();
    console.log('✅ Confirmed: deleted item not visible in feed');
  });

  // ── Teardown ──
  test.afterAll(async () => {
    console.log('\n--- Teardown ---');
    await kcHelper.deleteUserByEmail(testEmail);
    await dbHelper.cleanupTestUser(testEmail);
    await dbHelper.close();
    console.log('--- Done ---\n');
  });
});
