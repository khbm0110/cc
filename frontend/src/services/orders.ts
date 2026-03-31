import api from './api';
import type { Order, OrderFilters, PaginatedResponse } from '@/types';

export const ordersService = {
  async getOrders(filters: OrderFilters): Promise<PaginatedResponse<Order>> {
    const params = new URLSearchParams();
    if (filters.status) params.set('status', filters.status);
    if (filters.symbol) params.set('symbol', filters.symbol);
    if (filters.side) params.set('side', filters.side);
    if (filters.from_date) params.set('from_date', filters.from_date);
    if (filters.to_date) params.set('to_date', filters.to_date);
    params.set('page', filters.page.toString());
    params.set('limit', filters.limit.toString());
    return api.get<PaginatedResponse<Order>>(`/api/orders?${params.toString()}`);
  },

  async getOrder(id: number): Promise<Order> {
    return api.get<Order>(`/api/orders/${id}`);
  },

  async cancelOrder(id: number): Promise<Order> {
    return api.post<Order>(`/api/orders/${id}/cancel`);
  },

  async getOrderStats(): Promise<{
    total: number;
    filled: number;
    failed: number;
    pending: number;
    canceled: number;
  }> {
    return api.get('/api/orders/stats');
  },
};
