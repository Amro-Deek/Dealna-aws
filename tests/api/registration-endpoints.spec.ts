import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';
import * as crypto from 'crypto';

const dbHelper = new DatabaseHelper();
const kcHelper = new KeycloakHelper();

test.describe('Isolated Registration API Tests', () => {
  
  // Use parallel execution for these isolated tests
  test.describe.configure({ mode: 'parallel' });

  test.beforeAll(async () => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');
  });

  test.afterAll(async () => {
    await dbHelper.close();
  });

  test('POST /request-activation - Failure (Invalid Domain)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: `invalid_${Date.now()}@gmail.com` }
    });
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details.reason).toContain('invalid university domain');
  });

  test('POST /request-activation - Success', async ({ request }) => {
    const email = `req_success_${Date.now()}@student.birzeit.edu`;
    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email }
    });
    expect(response.status()).toBe(204);

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });

  test('POST /request-activation - Failure (Already Requested)', async ({ request }) => {
    const email = `req_fail_${Date.now()}@student.birzeit.edu`;
    // Seed database state directly
    await dbHelper.seedPendingPreRegistration(email, crypto.randomUUID());

    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email }
    });
    
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.details.reason).toContain('Activation already requested');

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });

  test('GET /activate - Success', async ({ request }) => {
    const email = `activate_success_${Date.now()}@student.birzeit.edu`;
    const token = crypto.randomUUID();
    // Seed pending pre-registration
    await dbHelper.seedPendingPreRegistration(email, token);

    const response = await request.get(`/api/v1/auth/student/activate?token=${token}`);
    expect(response.status()).toBe(204);

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });

  test('GET /activate - Failure (Already Verified)', async ({ request }) => {
    const email = `activate_fail_${Date.now()}@student.birzeit.edu`;
    // Seed verified pre-registration
    await dbHelper.seedVerifiedPreRegistration(email);
    
    // The seed method uses a specific token, we could fetch it, but let's fetch to be safe
    const token = await dbHelper.getStudentActivationToken(email);

    const response = await request.get(`/api/v1/auth/student/activate?token=${token}`);
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('already verified');

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });

  test('POST /complete - Success', async ({ request }) => {
    const email = `complete_success_${Date.now()}@student.birzeit.edu`;
    // Seed database so the user is verified and ready to complete
    await dbHelper.seedVerifiedPreRegistration(email);

    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: email,
        displayName: `Test Student ${Date.now()}`,
        password: 'StrongPassword123!',
        major: 'Computer Science',
        academicYear: 3,
      }
    });
    
    if (response.status() !== 204) {
      console.error(await response.text());
    }
    expect(response.status()).toBe(204);

    // Cleanup (both Keycloak and DB)
    await kcHelper.deleteUserByEmail(email);
    await dbHelper.cleanupTestUser(email);
  });

  test('POST /complete - Failure (Not Verified Yet)', async ({ request }) => {
    const email = `complete_fail_unverified_${Date.now()}@student.birzeit.edu`;
    // Seed pending (NOT verified)
    await dbHelper.seedPendingPreRegistration(email, crypto.randomUUID());

    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: email,
        displayName: `Test Student ${Date.now()}`,
        password: 'StrongPassword123!',
      }
    });
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('email not verified');

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });

  test('POST /complete - Failure (Already Completed)', async ({ request }) => {
    const email = `complete_fail_done_${Date.now()}@student.birzeit.edu`;
    // Seed already completed state
    await dbHelper.seedCompletedPreRegistration(email);

    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: email,
        displayName: `Test Student ${Date.now()}`,
        password: 'StrongPassword123!',
      }
    });
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('registration already completed');

    // Cleanup
    await dbHelper.cleanupTestUser(email);
  });
});
