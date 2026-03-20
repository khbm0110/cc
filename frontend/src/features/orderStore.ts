import { create } from 'zustand';
import type { Order, OrderFilters } from '@/types';
import { ordersService } from '@/services/orders';

interface OrderState {
  orders: Order[];
  totalOrders: number;
  totalPages: number;
  currentPage: number;
  isLoading: boolean;
  error: string | null;
  filters: OrderFilters;
  stats: { total: number; filled: number; failed: number; pending: number; canceled: number } | null;
  fetchOrders: (filters?: Partial<OrderFilters>) => Promise<void>;
  fetchStats: () => Promise<void>;
  cancelOrder: (id: number) => Promise<void>;
  setFilters: (filters: Partial<OrderFilters>) => void;
}

export const useOrderStore = create<OrderState>((set, get) => ({
  orders: [],
  totalOrders: 0,
  totalPages: 0,
  currentPage: 1,
  isLoading: false,
  error: null,
  filters: { page: 1, limit: 20 },
  stats: null,

  fetchOrders: async (newFilters) => {
    const filters = { ...get().filters, ...newFilters };
    set({ isLoading: true, error: null, filters });
    try {
      const data = await ordersService.getOrders(filters);
      set({
        orders: data.data,
        totalOrders: data.total,
        totalPages: data.total_pages,
        currentPage: data.page,
        isLoading: false,
      });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch orders', isLoading: false });
    }
  },

  fetchStats: async () => {
    try {
      const stats = await ordersService.getOrderStats();
      set({ stats });
    } catch {
      // silently fail for stats
    }
  },

  cancelOrder: async (id) => {
    try {
      await ordersService.cancelOrder(id);
      await get().fetchOrders();
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to cancel order' });
    }
  },

  setFilters: (newFilters) => {
    const filters = { ...get().filters, ...newFilters };
    set({ filters });
  },
}));
