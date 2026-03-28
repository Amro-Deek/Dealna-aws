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
}
