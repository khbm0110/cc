import api from './api';
import type { User, AuthTokens, LoginRequest, RegisterRequest } from '@/types';

export const authService = {
  async login(credentials: LoginRequest): Promise<AuthTokens & { user: User }> {
    const data = await api.post<AuthTokens & { user: User }>('/api/auth/login', credentials);
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
    }
    return data;
  },

  async register(data: RegisterRequest): Promise<AuthTokens & { user: User }> {
    const result = await api.post<AuthTokens & { user: User }>('/api/auth/register', data);
    if (typeof window !== 'undefined') {
      localStorage.setItem('access_token', result.access_token);
      localStorage.setItem('refresh_token', result.refresh_token);
    }
    return result;
  },

  async forgotPassword(email: string): Promise<{ message: string }> {
    return api.post('/api/auth/forgot-password', { email });
  },

  async getProfile(): Promise<User> {
    return api.get<User>('/api/auth/me');
  },

  async updateProfile(data: Partial<User>): Promise<User> {
    return api.put<User>('/api/auth/me', data);
  },

  async changePassword(oldPassword: string, newPassword: string): Promise<{ message: string }> {
    return api.post('/api/auth/change-password', { old_password: oldPassword, new_password: newPassword });
  },

  logout(): void {
    if (typeof window !== 'undefined') {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');
      window.location.href = '/auth/login';
    }
  },

  isAuthenticated(): boolean {
    if (typeof window === 'undefined') return false;
    return !!localStorage.getItem('access_token');
  },
};
