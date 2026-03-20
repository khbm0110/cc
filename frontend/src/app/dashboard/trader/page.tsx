'use client';

import React, { useEffect, useState } from 'react';
import Card, { StatCard } from '@/components/ui/Card';
import OrdersTable from '@/components/tables/OrdersTable';
import PerformanceChart from '@/components/charts/PerformanceChart';
import Modal from '@/components/ui/Modal';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import Button from '@/components/ui/Button';
import { useSignalStore } from '@/features/signalStore';
import { useOrderStore } from '@/features/orderStore';
import { formatCurrency, formatNumber, formatDateTime } from '@/utils/helpers';
import { StatusBadge } from '@/components/ui/Badge';
import type { ChartDataPoint, CreateSignalRequest } from '@/types';

const mockPerformanceData: ChartDataPoint[] = [
  { date: 'Mon', value: 12 },
  { date: 'Tue', value: 8 },
  { date: 'Wed', value: 15 },
  { date: 'Thu', value: 10 },
  { date: 'Fri', value: 18 },
  { date: 'Sat', value: 6 },
  { date: 'Sun', value: 14 },
];

const initialFormState: CreateSignalRequest = {
  symbol: '',
  side: 'BUY',
  quantity: 0,
  price: 0,
  stop_loss: undefined,
  take_profit: undefined,
  max_slippage: undefined,
};

export default function TraderDashboard() {
  const { signals, fetchSignals, createSignal, deleteSignal, performance, fetchPerformance, isLoading: signalsLoading } = useSignalStore();
  const { orders, fetchOrders, isLoading: ordersLoading } = useOrderStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [form, setForm] = useState<CreateSignalRequest>(initialFormState);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});

  useEffect(() => {
    fetchSignals();
    fetchOrders({ limit: 10 });
    fetchPerformance();
  }, [fetchSignals, fetchOrders, fetchPerformance]);

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};
    if (!form.symbol.trim()) errors.symbol = 'Symbol is required';
    if (form.quantity <= 0) errors.quantity = 'Quantity must be greater than 0';
    if (form.price <= 0) errors.price = 'Price must be greater than 0';
    if (form.stop_loss !== undefined && form.stop_loss <= 0) errors.stop_loss = 'Invalid stop loss';
    if (form.take_profit !== undefined && form.take_profit <= 0) errors.take_profit = 'Invalid take profit';
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateSignal = async () => {
    if (!validateForm()) return;
    try {
      await createSignal(form);
      setIsCreateModalOpen(false);
      setForm(initialFormState);
    } catch {
      // error in store
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Trader Dashboard</h1>
          <p className="text-gray-500 dark:text-gray-400 mt-1">Manage signals and track performance</p>
        </div>
        <Button onClick={() => setIsCreateModalOpen(true)}>
          <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
          </svg>
          New Signal
        </Button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Signals"
          value={performance?.total_signals ?? signals.length}
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" /></svg>}
        />
        <StatCard
          title="Win Rate"
          value={`${(performance?.win_rate ?? 71.8).toFixed(1)}%`}
          change="Last 30 days"
          changeType="positive"
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>}
        />
        <StatCard
          title="Avg Profit"
          value={formatCurrency(performance?.avg_profit ?? 245.30)}
          changeType="positive"
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>}
        />
        <StatCard
          title="Active Orders"
          value={orders.filter((o) => o.status === 'PENDING' || o.status === 'EXECUTING').length}
          icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>}
        />
      </div>

      {/* Performance Chart */}
      <Card>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Weekly Signal Performance</h3>
        <PerformanceChart data={mockPerformanceData} />
      </Card>

      {/* Signals List */}
      <Card>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Recent Signals</h3>
        {signalsLoading ? (
          <div className="animate-pulse space-y-3">
            {[...Array(3)].map((_, i) => <div key={i} className="h-12 bg-gray-200 dark:bg-gray-700 rounded" />)}
          </div>
        ) : signals.length === 0 ? (
          <p className="text-gray-500 dark:text-gray-400 text-center py-8">No signals created yet. Click &quot;New Signal&quot; to get started.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-200 dark:border-gray-700">
                  <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Signal ID</th>
                  <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Symbol</th>
                  <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Side</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Price</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Quantity</th>
                  <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Time</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Actions</th>
                </tr>
              </thead>
              <tbody>
                {signals.map((signal) => (
                  <tr key={signal.signal_id} className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="py-3 px-4 font-mono text-xs text-gray-600 dark:text-gray-400">{signal.signal_id.slice(0, 8)}...</td>
                    <td className="py-3 px-4 font-medium text-gray-900 dark:text-white">{signal.symbol}</td>
                    <td className="py-3 px-4">
                      <StatusBadge status={signal.side === 'BUY' ? 'FILLED' : 'FAILED'} />
                    </td>
                    <td className="py-3 px-4 text-right font-mono">{formatCurrency(signal.price)}</td>
                    <td className="py-3 px-4 text-right font-mono">{formatNumber(signal.quantity, 4)}</td>
                    <td className="py-3 px-4 text-xs text-gray-500">{signal.created_at ? formatDateTime(signal.created_at) : 'N/A'}</td>
                    <td className="py-3 px-4 text-right">
                      <button onClick={() => deleteSignal(signal.signal_id)} className="text-red-500 hover:text-red-700 text-xs font-medium">Delete</button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Recent Orders */}
      <Card>
        <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Recent Orders</h3>
        <OrdersTable orders={orders} isLoading={ordersLoading} />
      </Card>

      {/* Create Signal Modal */}
      <Modal isOpen={isCreateModalOpen} onClose={() => setIsCreateModalOpen(false)} title="Create Trade Signal" size="lg">
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Input label="Symbol" placeholder="BTCUSDT" value={form.symbol} onChange={(e) => setForm({ ...form, symbol: e.target.value.toUpperCase() })} error={formErrors.symbol} />
            <Select label="Side" options={[{ value: 'BUY', label: 'BUY' }, { value: 'SELL', label: 'SELL' }]} value={form.side} onChange={(e) => setForm({ ...form, side: e.target.value as 'BUY' | 'SELL' })} />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input label="Price" type="number" step="0.01" value={form.price || ''} onChange={(e) => setForm({ ...form, price: parseFloat(e.target.value) || 0 })} error={formErrors.price} />
            <Input label="Quantity" type="number" step="0.0001" value={form.quantity || ''} onChange={(e) => setForm({ ...form, quantity: parseFloat(e.target.value) || 0 })} error={formErrors.quantity} />
          </div>
          <div className="grid grid-cols-3 gap-4">
            <Input label="Stop Loss" type="number" step="0.01" placeholder="Optional" value={form.stop_loss || ''} onChange={(e) => setForm({ ...form, stop_loss: parseFloat(e.target.value) || undefined })} error={formErrors.stop_loss} />
            <Input label="Take Profit" type="number" step="0.01" placeholder="Optional" value={form.take_profit || ''} onChange={(e) => setForm({ ...form, take_profit: parseFloat(e.target.value) || undefined })} error={formErrors.take_profit} />
            <Input label="Max Slippage %" type="number" step="0.1" placeholder="Optional" value={form.max_slippage || ''} onChange={(e) => setForm({ ...form, max_slippage: parseFloat(e.target.value) || undefined })} />
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setIsCreateModalOpen(false)}>Cancel</Button>
            <Button onClick={handleCreateSignal} isLoading={signalsLoading}>Create Signal</Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
