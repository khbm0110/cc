import api from './api';
import type { User, Plan, SystemMetrics, ReconcilerLog, PaginatedResponse } from '@/types';

export const adminService = {
  // User management
  async getUsers(page = 1, limit = 20): Promise<PaginatedResponse<User>> {
    return api.get<PaginatedResponse<User>>(`/api/admin/users?page=${page}&limit=${limit}`);
  },

  async updateUser(id: number, data: Partial<User>): Promise<User> {
    return api.put<User>(`/api/admin/users/${id}`, data);
  },

  async suspendUser(id: number): Promise<{ message: string }> {
    return api.post(`/api/admin/users/${id}/suspend`);
  },

  async deleteUser(id: number): Promise<{ message: string }> {
    return api.delete(`/api/admin/users/${id}`);
  },

  // Plan management
  async getPlans(): Promise<Plan[]> {
    return api.get<Plan[]>('/api/admin/plans');
  },

  async createPlan(data: Omit<Plan, 'id' | 'created_at' | 'updated_at'>): Promise<Plan> {
    return api.post<Plan>('/api/admin/plans', data);
  },

  async updatePlan(id: number, data: Partial<Plan>): Promise<Plan> {
    return api.put<Plan>(`/api/admin/plans/${id}`, data);
  },

  async deletePlan(id: number): Promise<{ message: string }> {
    return api.delete(`/api/admin/plans/${id}`);
  },

  // Metrics
  async getSystemMetrics(): Promise<SystemMetrics> {
    return api.get<SystemMetrics>('/api/admin/metrics');
  },

  // Reconciler logs
  async getReconcilerLogs(page = 1, limit = 50): Promise<PaginatedResponse<ReconcilerLog>> {
    return api.get<PaginatedResponse<ReconcilerLog>>(`/api/admin/reconciler/logs?page=${page}&limit=${limit}`);
  },
};
