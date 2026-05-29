import * as dotenv from 'dotenv';
import * as path from 'path';

dotenv.config({ path: path.resolve(__dirname, '../.env.test') });

/**
 * Keycloak Helper
 * Interacts directly with the Keycloak Admin REST API to clean up test users.
 */
export class KeycloakHelper {
  private baseURL: string;
  private realm: string;
  private adminClient: string;
  private adminSecret: string;

  constructor() {
    // Trim trailing slash if present
    const rawURL = process.env.KEYCLOAK_BASE_URL || 'http://localhost:8080';
    this.baseURL = rawURL.replace(/\/$/, '');
    
    this.realm = process.env.KEYCLOAK_REALM || 'Dealna';
    this.adminClient = process.env.KEYCLOAK_ADMIN_CLIENT_ID || 'dealna-backend';
    this.adminSecret = process.env.KEYCLOAK_ADMIN_CLIENT_SECRET || 'HtcTRtda0F53HXbf3uIUpTqakX2albXQ';
  }

  /**
   * Fetches an admin access token using client credentials from the target realm
   */
  private async getAdminToken(): Promise<string> {
    const tokenEndpoint = `${this.baseURL}/realms/${this.realm}/protocol/openid-connect/token`;
    
    const params = new URLSearchParams();
    params.append('client_id', this.adminClient);
    params.append('client_secret', this.adminSecret);
    params.append('grant_type', 'client_credentials');

    const response = await fetch(tokenEndpoint, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/x-www-form-urlencoded',
      },
      body: params.toString(),
    });

    if (!response.ok) {
      throw new Error(`Failed to get Keycloak admin token: ${await response.text()}`);
    }

    const data = await response.json();
    return data.access_token;
  }

  async ensureRoleExists(roleName: string): Promise<void> {
    try {
      const token = await this.getAdminToken();
      const rolesEndpoint = `${this.baseURL}/admin/realms/${this.realm}/roles`;
      
      const checkRes = await fetch(`${rolesEndpoint}/${roleName}`, {
        headers: { Authorization: `Bearer ${token}` }
      });
      if (checkRes.ok) return;

      const createRes = await fetch(rolesEndpoint, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ name: roleName })
      });
      if (!createRes.ok && createRes.status !== 409) {
        throw new Error(`Failed to create role ${roleName}: ${await createRes.text()}`);
      }
    } catch (e) {
      console.error('ensureRoleExists error:', e);
    }
  }

  /**
   * Deletes a user by their email address
   */
  async deleteUserByEmail(email: string): Promise<void> {
    try {
      const token = await this.getAdminToken();

      // 1. Search for user by email
      const searchUrl = `${this.baseURL}/admin/realms/${this.realm}/users?email=${encodeURIComponent(email)}&exact=true`;
      const searchRes = await fetch(searchUrl, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!searchRes.ok) {
        throw new Error(`Failed to search user in Keycloak: ${await searchRes.text()}`);
      }

      const users = await searchRes.json();
      if (!users || users.length === 0) {
        console.log(`ℹ️ Keycloak user ${email} not found. Skipping cleanup.`);
        return;
      }

      const userId = users[0].id;

      // 2. Delete user by ID
      const deleteUrl = `${this.baseURL}/admin/realms/${this.realm}/users/${userId}`;
      const deleteRes = await fetch(deleteUrl, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      });

      if (!deleteRes.ok) {
        throw new Error(`Failed to delete user in Keycloak: ${await deleteRes.text()}`);
      }

      console.log(`🧹 Cleaned up Keycloak user: ${email}`);
    } catch (e) {
      console.error(`❌ Error cleaning up Keycloak user ${email}:`, e);
    }
  }

  /**
   * Forcefully verifies a user's email so they can log in
   */
  async verifyUserEmail(email: string): Promise<void> {
    try {
      const token = await this.getAdminToken();

      const searchUrl = `${this.baseURL}/admin/realms/${this.realm}/users?email=${encodeURIComponent(email)}&exact=true`;
      const searchRes = await fetch(searchUrl, {
        headers: { 'Authorization': `Bearer ${token}` }
      });

      if (!searchRes.ok) throw new Error(`Failed to search user`);

      const users = await searchRes.json();
      if (!users || users.length === 0) {
        throw new Error(`User ${email} not found in Keycloak for verification`);
      }

      const userId = users[0].id;
      const userObj = users[0];
      userObj.emailVerified = true;
      userObj.requiredActions = [];

      const updateUrl = `${this.baseURL}/admin/realms/${this.realm}/users/${userId}`;
      const updateRes = await fetch(updateUrl, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(userObj),
      });

      if (!updateRes.ok) throw new Error(`Failed to update user: ${await updateRes.text()}`);

      console.log(`✅ Verified Keycloak user email: ${email}`);
    } catch (e) {
      console.error(`❌ Error verifying Keycloak user ${email}:`, e);
    }
  }
}
