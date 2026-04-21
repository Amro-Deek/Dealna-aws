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

test.describe.serial('Profile & Follow Integration API', () => {
  const dbHelper = new DatabaseHelper();
  const kcHelper = new KeycloakHelper();

  const ts = Date.now();
  const userA = {
    email: `usera_${ts}@student.birzeit.edu`,
    display: `User A ${ts}`,
    pass: 'Pass123!',
    token: '',
    profileId: '',
  };
  const userB = {
    email: `userb_${ts}@student.birzeit.edu`,
    display: `User B ${ts}`,
    pass: 'Pass123!',
    token: '',
    profileId: '',
  };

  test.beforeAll(async ({ request }) => {
    await dbHelper.ensureUniversityExists('Birzeit University', 'birzeit.edu');

    // Register User A & User B
    for (const u of [userA, userB]) {
      // 1. Request
      await request.post('/api/v1/auth/student/request-activation', { data: { email: u.email } });
      const actToken = await dbHelper.getStudentActivationToken(u.email);
      // 2. Activate
      await request.get(`/api/v1/auth/student/activate?token=${actToken}`);
      // 3. Complete
      await request.post('/api/v1/auth/student/complete', {
        data: {
          email: u.email,
          displayName: u.display,
          password: u.pass,
          major: 'CS',
          academicYear: 1,
        },
      });
      // 4. Token
      u.token = await getStudentToken(u.email, u.pass);
    }
  });

  test.afterAll(async () => {
    await dbHelper.cleanupTestUser(userA.email);
    await dbHelper.cleanupTestUser(userB.email);
    try { await kcHelper.deleteUserByEmail(userA.email); } catch {}
    try { await kcHelper.deleteUserByEmail(userB.email); } catch {}
    await dbHelper.close();
  });

  test('User A should be able to get their own profile_id', async ({ request }) => {
    const res = await request.get('/api/v1/profile', {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('profile_id');
    expect(body.profile_id).toBeTruthy();
    userA.profileId = body.profile_id;
  });

  test('User B should be able to get their own profile_id', async ({ request }) => {
    const res = await request.get('/api/v1/profile', {
      headers: { Authorization: `Bearer ${userB.token}` },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body).toHaveProperty('profile_id');
    userB.profileId = body.profile_id;
  });

  test('User A should be able to view User B public profile', async ({ request }) => {
    const res = await request.get(`/api/v1/users/${userB.profileId}/profile`, {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    expect(res.status()).toBe(200);
    const body = await res.json();
    expect(body.display_name).toBe(userB.display);
    expect(body.profile_id).toBe(userB.profileId);
    expect(body).toHaveProperty('follower_count');
    expect(body).toHaveProperty('following_count');
  });

  test('User A follows User B successfully', async ({ request }) => {
    const res = await request.post(`/api/v1/users/${userB.profileId}/follow`, {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    expect(res.status()).toBe(204);

    // Verify IsFollowing
    const checkRes = await request.get(`/api/v1/users/${userB.profileId}/is-following`, {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    expect(checkRes.status()).toBe(200);
    const checkBody = await checkRes.json();
    expect(checkBody.is_following).toBe(true);
  });

  test('Profiles should reflect updated following/follower counts', async ({ request }) => {
    // Check User A (Following 1)
    const resA = await request.get('/api/v1/profile', {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    const bodyA = await resA.json();
    expect(bodyA.following_count).toBe(1);

    // Check User B (Follower 1)
    const resB = await request.get('/api/v1/profile', {
      headers: { Authorization: `Bearer ${userB.token}` },
    });
    const bodyB = await resB.json();
    expect(bodyB.follower_count).toBe(1);
  });

  test('User A unfollows User B successfully', async ({ request }) => {
    const res = await request.delete(`/api/v1/users/${userB.profileId}/unfollow`, {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    expect(res.status()).toBe(204);

    // Counts should go back to 0
    const resA = await request.get('/api/v1/profile', {
      headers: { Authorization: `Bearer ${userA.token}` },
    });
    const bodyA = await resA.json();
    expect(bodyA.following_count).toBe(0);
  });
});
