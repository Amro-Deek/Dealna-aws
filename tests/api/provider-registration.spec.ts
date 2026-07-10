import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

test.describe.serial('Provider Registration API Flow', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const timestamp = Date.now();
  const testEmail = `provider_${timestamp}@dealna.com`;
  const password = 'StrongPassword123!';

	let accessToken: string | null = null;
	let nationalIdObjectKey: string | null = null;
	let proofObjectKey: string | null = null;
	let universityId: string | null = null;
	let applicantId: string | null = null;
	const adminEmail = `admin_${timestamp}@dealna.com`;
	let adminToken: string | null = null;

  test.beforeAll(async () => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');
    const bzu = await dbHelper.getPool().query(`SELECT university_id FROM university WHERE domain = 'birzeit.edu' LIMIT 1`);
    if (bzu.rows.length > 0) {
      universityId = bzu.rows[0].university_id;
    } else {
      throw new Error("Birzeit University not found in dbHelper");
    }
  });

  test('POST /api/v1/auth/providers/request-activation - Success', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/request-activation', {
      data: {
        email: testEmail
      }
    });
    
    if (response.status() !== 204) {
      console.error(await response.text());
    }
    expect(response.status()).toBe(204); // No Content
  });

  test('Fetch token from Database for activation', async ({ request }) => {
    const res = await dbHelper.getPool().query('SELECT token FROM provider_pre_registration WHERE email = $1', [testEmail]);
    expect(res.rows.length).toBe(1);
    const token = res.rows[0].token;

    // Call activate
    const actRes = await request.get(`/api/v1/auth/providers/activate?token=${token}`);
    expect(actRes.status()).toBe(204);
  });

  test('POST /api/v1/auth/providers/complete - Success', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/complete', {
      data: {
        email: testEmail,
        password: password
      }
    });
    
    if (response.status() !== 200) {
      console.error(await response.text());
    }
    expect(response.status()).toBe(200); // OK (returns tokens)
  });

  test('POST /api/v1/auth/login - Success (Get APPLICANT Token)', async ({ request }) => {
    let response;
    let body;
    
    // Retry login up to 5 times (Keycloak sometimes takes a moment to sync passwords/verification)
    for (let i = 0; i < 5; i++) {
      response = await request.post('/api/v1/auth/login', {
        data: {
          email: testEmail,
          password: password
        }
      });
      
      if (response.status() === 200) {
        break;
      }
      
      console.log(`Login attempt ${i + 1} failed with status ${response.status()}:`, await response.text());
      await new Promise(resolve => setTimeout(resolve, 1000));
    }
    
    if (response.status() !== 200) {
      console.error(await response.text());
    }
    expect(response.status(), `Login finally failed with status ${response.status()}`).toBe(200);
    body = await response.json();
    expect(body.access_token).toBeDefined();
    expect(body.user.role).toBe('APPLICANT');
    accessToken = body.access_token;
  });

  test('POST /api/v1/auth/providers/application/start - Draft Application', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/application/start', {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        university_id: universityId,
        business_name: 'Test Provider Store',
        phone_number: '0599123456',
        business_type: 'Retail',
        address: 'Ramallah'
      }
    });

    if (response.status() !== 200) {
      console.error(await response.text());
    }
    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.business_name).toBe('Test Provider Store');
    expect(body.status).toBe('DRAFT');
  });

  test('POST /api/v1/auth/providers/application/document-url - NATIONAL_ID', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/application/document-url', {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        document_type: 'NATIONAL_ID',
        original_filename: 'id.pdf',
        content_type: 'application/pdf'
      }
    });

    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.url).toContain('https://');
    expect(body.object_key).toContain('NATIONAL_ID');
    nationalIdObjectKey = body.object_key;
  });

  test('POST /api/v1/auth/providers/application/document-url - PROOF_OF_OWNERSHIP', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/application/document-url', {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        document_type: 'PROOF_OF_OWNERSHIP',
        original_filename: 'proof.pdf',
        content_type: 'application/pdf'
      }
    });

    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.url).toContain('https://');
    expect(body.object_key).toContain('PROOF_OF_OWNERSHIP');
    proofObjectKey = body.object_key;
  });

  test('POST /api/v1/auth/providers/application/submit - Failure (Missing Documents)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/application/submit', {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(400); // Bad Request (Missing docs)
    const body = await response.json();
    expect(body.message).toContain('Validation failed');
    expect(body.details.reason).toContain('NATIONAL_ID and PROOF_OF_OWNERSHIP are both required');
  });

  test('POST /api/v1/auth/providers/application/document/confirm - Upload Docs', async ({ request }) => {
    // Confirm National ID
    let response = await request.post('/api/v1/auth/providers/application/document/confirm', {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        object_key: nationalIdObjectKey,
        document_type: 'NATIONAL_ID',
        original_filename: 'id.pdf',
        content_type: 'application/pdf',
        size_bytes: 1024
      }
    });
    expect(response.status()).toBe(204);

    // Confirm Proof of Ownership
    response = await request.post('/api/v1/auth/providers/application/document/confirm', {
      headers: { Authorization: `Bearer ${accessToken}` },
      data: {
        object_key: proofObjectKey,
        document_type: 'PROOF_OF_OWNERSHIP',
        original_filename: 'proof.pdf',
        content_type: 'application/pdf',
        size_bytes: 2048
      }
    });
    expect(response.status()).toBe(204);
  });

  test('POST /api/v1/auth/providers/application/submit - Success (Pending Review)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/providers/application/submit', {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(204); // Success!
  });

  test('GET /api/v1/auth/providers/application/status - Returns PENDING_REVIEW', async ({ request }) => {
    const response = await request.get('/api/v1/auth/providers/application/status', {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.status).toBe('PENDING_REVIEW');
    applicantId = body.applicant_id;
  });

  test('Admin Registration & Login', async ({ request }) => {
    // 1. Register admin as a normal user
    // 1. Request Activation
    const resReg = await request.post('/api/v1/auth/providers/request-activation', {
      data: {
        email: adminEmail
      }
    });
    expect(resReg.status()).toBe(204);

    // 2. Fetch token from Database for activation
    const resAdminToken = await dbHelper.getPool().query('SELECT token FROM provider_pre_registration WHERE email = $1', [adminEmail]);
    const adminTokenVal = resAdminToken.rows[0].token;
    const actAdminRes = await request.get(`/api/v1/auth/providers/activate?token=${adminTokenVal}`);
    expect(actAdminRes.status()).toBe(204);

    // 3. Complete Registration
    const completeAdminRes = await request.post('/api/v1/auth/providers/complete', {
      data: {
        email: adminEmail,
        password: 'Password123!'
      }
    });
    expect(completeAdminRes.status()).toBe(200);

    // (Removed manual verifyUserEmail because backend already sets emailVerified=true during registration)

    // 3. Promote to ADMIN in Postgres and add to admin table
    const resUpdate = await dbHelper.getPool().query(`UPDATE "User" SET role = 'ADMIN' WHERE email = $1 RETURNING user_id`, [adminEmail]);
    const adminUserId = resUpdate.rows[0].user_id;
    await dbHelper.getPool().query(`INSERT INTO admin (user_id, admin_name) VALUES ($1, 'Test Admin') ON CONFLICT (user_id) DO NOTHING`, [adminUserId]);

    // 4. Login to get Admin Token
    const resLog = await request.post('/api/v1/auth/login', {
      data: {
        email: adminEmail,
        password: 'Password123!'
      }
    });
    expect(resLog.status()).toBe(200);
    const body = await resLog.json();
    adminToken = body.access_token;
    expect(body.user.role).toBe('ADMIN');
  });

  test('POST /api/v1/admin/providers/{id}/reject - Rejects with comment', async ({ request }) => {
    const response = await request.post(`/api/v1/admin/providers/${applicantId}/reject`, {
      headers: { Authorization: `Bearer ${adminToken}` },
      data: { comment: 'Need better documentation' }
    });

    expect(response.status()).toBe(204);
  });

  test('GET /api/v1/auth/providers/application/status - Returns REJECTED', async ({ request }) => {
    const response = await request.get('/api/v1/auth/providers/application/status', {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.status).toBe('REJECTED');
  });

  test('POST /api/v1/admin/providers/{id}/approve - Approves Application', async ({ request }) => {
    // Manually set back to PENDING_REVIEW for testing approve
    await dbHelper.getPool().query(`UPDATE providerapplication SET status = 'PENDING_REVIEW' WHERE applicant_id = $1`, [applicantId]);

    const response = await request.post(`/api/v1/admin/providers/${applicantId}/approve`, {
      headers: { Authorization: `Bearer ${adminToken}` }
    });

    if (response.status() !== 204) {
      console.error('Approve failed:', await response.text());
    }
    expect(response.status()).toBe(204);
  });

  test('GET /api/v1/auth/providers/application/status - Returns APPROVED', async ({ request }) => {
    const response = await request.get('/api/v1/auth/providers/application/status', {
      headers: { Authorization: `Bearer ${accessToken}` }
    });

    expect(response.status()).toBe(200);
    const body = await response.json();
    expect(body.status).toBe('APPROVED');
  });

  test.beforeAll(async () => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');
    await dbHelper.cleanupTestUser(testEmail);
    await kcHelper.deleteUserByEmail(testEmail);
    await dbHelper.cleanupTestUser(adminEmail);
    await kcHelper.deleteUserByEmail(adminEmail);
  });

  test.afterAll(async () => {
    console.log(`\n--- Starting Teardown ---`);
    await kcHelper.deleteUserByEmail(testEmail);
    await dbHelper.cleanupTestUser(testEmail);
    await kcHelper.deleteUserByEmail(adminEmail);
    await dbHelper.cleanupTestUser(adminEmail);
    await dbHelper.close();
    console.log(`--- Teardown Complete ---\n`);
  });
});
