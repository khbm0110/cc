import { create } from 'zustand';
import type { TradeSignal, CreateSignalRequest } from '@/types';
import { signalsService } from '@/services/signals';

interface SignalState {
  signals: TradeSignal[];
  totalSignals: number;
  isLoading: boolean;
  error: string | null;
  performance: {
    total_signals: number;
    successful: number;
    failed: number;
    win_rate: number;
    avg_profit: number;
  } | null;
  fetchSignals: (page?: number) => Promise<void>;
  createSignal: (data: CreateSignalRequest) => Promise<void>;
  deleteSignal: (signalId: string) => Promise<void>;
  fetchPerformance: () => Promise<void>;
}

export const useSignalStore = create<SignalState>((set, get) => ({
  signals: [],
  totalSignals: 0,
  isLoading: false,
  error: null,
  performance: null,

  fetchSignals: async (page = 1) => {
    set({ isLoading: true, error: null });
    try {
      const data = await signalsService.getSignals(page);
      set({ signals: data.data, totalSignals: data.total, isLoading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to fetch signals', isLoading: false });
    }
  },

  createSignal: async (data) => {
    set({ isLoading: true, error: null });
    try {
      await signalsService.createSignal(data);
      await get().fetchSignals();
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to create signal', isLoading: false });
      throw err;
    }
  },

  deleteSignal: async (signalId) => {
    try {
      await signalsService.deleteSignal(signalId);
      await get().fetchSignals();
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to delete signal' });
    }
  },

  fetchPerformance: async () => {
    try {
      const performance = await signalsService.getSignalPerformance();
      set({ performance });
    } catch {
      // silently fail
    }
  },
}));
