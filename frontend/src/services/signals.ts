import api from './api';
import type { TradeSignal, CreateSignalRequest, PaginatedResponse } from '@/types';

export const signalsService = {
  async getSignals(page = 1, limit = 20): Promise<PaginatedResponse<TradeSignal>> {
    return api.get<PaginatedResponse<TradeSignal>>(`/api/signals?page=${page}&limit=${limit}`);
  },

  async createSignal(data: CreateSignalRequest): Promise<TradeSignal> {
    return api.post<TradeSignal>('/api/signals', data);
  },

  async deleteSignal(signalId: string): Promise<{ message: string }> {
    return api.delete(`/api/signals/${signalId}`);
  },

  async getSignalPerformance(): Promise<{
    total_signals: number;
    successful: number;
    failed: number;
    win_rate: number;
    avg_profit: number;
  }> {
    return api.get('/api/signals/performance');
  },
};
