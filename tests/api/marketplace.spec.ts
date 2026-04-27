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

// ─────────────────────────────────────────────────────────
// Minimal valid JPEG (smallest possible — works for S3 upload validation)
function createMinimalJpeg(): Buffer {
  // Smallest valid JPEG: SOI + APP0 (JFIF) + minimal frame + EOI
  return Buffer.from([
    0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01,
    0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0xFF, 0xDB, 0x00, 0x43,
    0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05, 0x08, 0x07, 0x07, 0x07, 0x09,
    0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D, 0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12,
    0x13, 0x0F, 0x14, 0x1D, 0x1A, 0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20,
    0x24, 0x2E, 0x27, 0x20, 0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29,
    0x2C, 0x30, 0x31, 0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32,
    0x3C, 0x2E, 0x33, 0x34, 0x32, 0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01,
    0x00, 0x01, 0x01, 0x01, 0x11, 0x00, 0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00,
    0x01, 0x05, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
    0x09, 0x0A, 0x0B, 0xFF, 0xC4, 0x00, 0xB5, 0x10, 0x00, 0x02, 0x01, 0x03,
    0x03, 0x02, 0x04, 0x03, 0x05, 0x05, 0x04, 0x04, 0x00, 0x00, 0x01, 0x7D,
    0x01, 0x02, 0x03, 0x00, 0x04, 0x11, 0x05, 0x12, 0x21, 0x31, 0x41, 0x06,
    0x13, 0x51, 0x61, 0x07, 0x22, 0x71, 0x14, 0x32, 0x81, 0x91, 0xA1, 0x08,
    0x23, 0x42, 0xB1, 0xC1, 0x15, 0x52, 0xD1, 0xF0, 0x24, 0x33, 0x62, 0x72,
    0x82, 0x09, 0x0A, 0x16, 0x17, 0x18, 0x19, 0x1A, 0x25, 0x26, 0x27, 0x28,
    0x29, 0x2A, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3A, 0x43, 0x44, 0x45,
    0x46, 0x47, 0x48, 0x49, 0x4A, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59,
    0x5A, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6A, 0x73, 0x74, 0x75,
    0x76, 0x77, 0x78, 0x79, 0x7A, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89,
    0x8A, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9A, 0xA2, 0xA3,
    0xA4, 0xA5, 0xA6, 0xA7, 0xA8, 0xA9, 0xAA, 0xB2, 0xB3, 0xB4, 0xB5, 0xB6,
    0xB7, 0xB8, 0xB9, 0xBA, 0xC2, 0xC3, 0xC4, 0xC5, 0xC6, 0xC7, 0xC8, 0xC9,
    0xCA, 0xD2, 0xD3, 0xD4, 0xD5, 0xD6, 0xD7, 0xD8, 0xD9, 0xDA, 0xE1, 0xE2,
    0xE3, 0xE4, 0xE5, 0xE6, 0xE7, 0xE8, 0xE9, 0xEA, 0xF1, 0xF2, 0xF3, 0xF4,
    0xF5, 0xF6, 0xF7, 0xF8, 0xF9, 0xFA, 0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01,
    0x00, 0x00, 0x3F, 0x00, 0x7B, 0x94, 0x11, 0x00, 0x00, 0x00, 0x00, 0x00,
    0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xD9,
  ]);
}

// ─────────────────────────────────────────────────────────
// Helper: PUT image buffer directly to S3 presigned URL
// ─────────────────────────────────────────────────────────
async function uploadToS3(presignedUrl: string): Promise<void> {
  const buffer = createMinimalJpeg();
  const res = await fetch(presignedUrl, {
    method: 'PUT',
    headers: { 'Content-Type': 'image/jpeg' },
    body: new Uint8Array(buffer),
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

    await uploadToS3(body.upload_url);
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

  // ── 8. Feed — university scoped (Excludes Self) ──
  test('GET /items/feed — 200 excludes newly created item (own items hidden)', async ({ request }) => {
    const res = await request.get('/api/v1/items/feed?limit=50', {
      headers: { Authorization: `Bearer ${accessToken}` },
    });

    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(Array.isArray(body)).toBe(true);

    // Now it SHOULD be missing because we hide our own items
    const found = body.find((i: any) => i.id === createdItemId);
    expect(found).toBeFalsy(); 
    console.log(`📰 Feed: ${body.length} item(s). Our item correctly hidden from self ✅`);
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
