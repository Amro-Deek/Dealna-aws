import { test, expect } from '@playwright/test';
import { DatabaseHelper } from '../utils/db';
import { KeycloakHelper } from '../utils/keycloak';

test.describe.serial('Student Registration API Flow', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();
  
  // Use a unique timestamp to generate dynamic test data and avoid collisions
  const timestamp = Date.now();
  const testEmail = `${timestamp}@student.birzeit.edu`;
  const invalidEmail = `test_${timestamp}@gmail.com`;
  const displayName = `Amro ${timestamp}`;
  const password = 'StrongPassword123!';
  
  let activationToken: string | null = null;

  test('POST /request-activation - Failure (Invalid Domain)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: invalidEmail }
    });
    
    expect(response.status()).toBe(400); // Bad Request (Validation Error)
    const body = await response.json();
    expect(body.code).toBe('VALIDATION_FAILED');
    expect(body.message).toBe('Validation failed');
    expect(body.details.field).toBe('email');
    expect(body.details.reason).toContain('invalid university domain');
  });

  test('POST /request-activation - Success', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: testEmail }
    });
    
    expect(response.status()).toBe(204); // No Content
  });

  test('POST /request-activation - Failure (Already Requested / Email in Use)', async ({ request }) => {
    // Wait for a short moment to ensure DB transaction is fully committed
    await new Promise((resolve) => setTimeout(resolve, 500));

    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: testEmail }
    });
    
    // Should fail because it's already requested
    expect(response.status()).toBe(400);
    const body = await response.json();
    expect(body.code).toBe('VALIDATION_FAILED');
    expect(body.message).toBe('Validation failed');
    expect(body.details.field).toBe('email');
    expect(body.details.reason).toContain('Activation already requested');
  });

  test('POST /resend - Success', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/resend', {
      data: { email: testEmail }
    });
    
    expect(response.status()).toBe(204); // No Content
  });

  test('Fetch token from Database for activation', async () => {
    // Sleep briefly to ensure the resend update is fully flushed to DB
    await new Promise(resolve => setTimeout(resolve, 1000));
    
    activationToken = await dbHelper.getStudentActivationToken(testEmail);
    expect(activationToken).not.toBeNull();
    // Validate UUID format roughly
    expect(activationToken).toMatch(/^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i);
  });

  test('GET /activate - Failure (Invalid Token)', async ({ request }) => {
    const fakeToken = '00000000-0000-0000-0000-000000000000';
    const response = await request.get(`/api/v1/auth/student/activate?token=${fakeToken}`);
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('invalid token');
  });

  test('GET /activate - Success', async ({ request }) => {
    // We know activationToken is set from the previous step
    const response = await request.get(`/api/v1/auth/student/activate?token=${activationToken}`);
    
    expect(response.status()).toBe(204); // No Content
  });

  test('GET /activate - Failure (Already Verified)', async ({ request }) => {
    const response = await request.get(`/api/v1/auth/student/activate?token=${activationToken}`);
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('already verified');
  });

  test('POST /complete - Failure (Invalid Domain in Body)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: invalidEmail, // Mismatch domain
        displayName,
        password,
        major: 'Computer Science',
        academicYear: 3,
      }
    });
    
    expect(response.status()).toBe(401); // Unauthorized because activation not requested for this email
    const body = await response.json();
    expect(body.message).toContain('activation not requested');
  });

  test('POST /complete - Success', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: testEmail,
        displayName,
        password,
        major: 'Computer Science',
        academicYear: 3,
      }
    });
    
    // Check if the response failed due to Keycloak issues or DB issues.
    // Assuming 204 No Content on true success
    if (response.status() !== 204) {
      console.error(await response.text());
    }
    expect(response.status()).toBe(204);
  });

  test('POST /complete - Failure (Already Completed)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/complete', {
      data: {
        email: testEmail,
        displayName,
        password,
      }
    });
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('registration already completed');
  });

  test('POST /resend - Failure (Already Completed)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/resend', {
      data: { email: testEmail }
    });
    
    expect(response.status()).toBe(401);
    const body = await response.json();
    expect(body.message).toContain('registration already completed');
  });

  test('POST /request-activation - Failure (Already a User)', async ({ request }) => {
    const response = await request.post('/api/v1/auth/student/request-activation', {
      data: { email: testEmail }
    });
    
    expect(response.status()).toBe(409); // Email Already Used
    const body = await response.json();
    expect(body.message).toContain('already in use');
  });

  test.afterAll(async () => {
    console.log(`\n--- Starting Teardown ---`);
    await kcHelper.deleteUserByEmail(testEmail);
    await dbHelper.cleanupTestUser(testEmail);
    console.log(`--- Teardown Complete ---\n`);
  });
});
