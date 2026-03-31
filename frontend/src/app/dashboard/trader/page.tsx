'use client';

import React, { useEffect, useState } from 'react';
import Card, { StatCard } from '@/components/ui/Card';
import OrdersTable from '@/components/tables/OrdersTable';
import Modal from '@/components/ui/Modal';
import Input from '@/components/ui/Input';
import Select from '@/components/ui/Select';
import Button from '@/components/ui/Button';
import { useSignalStore } from '@/features/signalStore';
import { useOrderStore } from '@/features/orderStore';
import { formatCurrency, formatNumber, formatDateTime } from '@/utils/helpers';
import { StatusBadge } from '@/components/ui/Badge';
import type { CreateSignalRequest } from '@/types';
import { FaCoins, FaChartLine, FaWallet, FaExchangeAlt, FaTrophy, FaCrown } from 'react-icons/fa';

interface Plan {
  id: number;
  name: string;
  virtual_balance: number;
  subscription_price: number;
  min_investment: number;
  description: string;
  trader_id: number | null;
}

interface VirtualPortfolio {
  symbol: string;
  quantity: number;
  avg_buy_price: number;
  current_value: number;
  profit_loss: number;
  profit_loss_pct: number;
}

const mockPlans: Plan[] = [
  {
    id: 1,
    name: 'Basic',
    virtual_balance: 3000,
    subscription_price: 0,
    min_investment: 1000,
    description: 'Plan with $3000 virtual balance',
    trader_id: 1
  },
  {
    id: 2,
    name: 'Pro',
    virtual_balance: 5000,
    subscription_price: 49,
    min_investment: 3000,
    description: 'Pro plan with $5000 virtual balance',
    trader_id: 1
  },
  {
    id: 3,
    name: 'Enterprise',
    virtual_balance: 10000,
    subscription_price: 199,
    min_investment: 10000,
    description: 'Enterprise plan with $10000 virtual balance',
    trader_id: 1
  }
];

const mockPortfolio: VirtualPortfolio[] = [
  { symbol: 'BTC', quantity: 0.05, avg_buy_price: 42000, current_value: 2150, profit_loss: 150, profit_loss_pct: 7.5 },
  { symbol: 'ETH', quantity: 2.5, avg_buy_price: 2200, current_value: 5750, profit_loss: 250, profit_loss_pct: 4.5 },
  { symbol: 'SOL', quantity: 50, avg_buy_price: 95, current_value: 5000, profit_loss: 250, profit_loss_pct: 5.26 },
];

const initialFormState: CreateSignalRequest = {
  symbol: '',
  side: 'BUY',
  quantity: 0,
  price: 0,
};

export default function TraderDashboard() {
  const { signals, fetchSignals, createSignal, deleteSignal, isLoading: signalsLoading } = useSignalStore();
  const { orders, fetchOrders, isLoading: ordersLoading } = useOrderStore();
  const [isCreateModalOpen, setIsCreateModalOpen] = useState(false);
  const [form, setForm] = useState<CreateSignalRequest>(initialFormState);
  const [formErrors, setFormErrors] = useState<Record<string, string>>({});
  const [selectedPlan, setSelectedPlan] = useState<Plan>(mockPlans[0]);
  const [portfolio] = useState<VirtualPortfolio[]>(mockPortfolio);
  const [virtualBalance, setVirtualBalance] = useState(selectedPlan.virtual_balance);
  const [isDark] = useState(true);

  useEffect(() => {
    fetchSignals();
    fetchOrders({ limit: 10 });
  }, [fetchSignals, fetchOrders]);

  useEffect(() => {
    // Update virtual balance when plan changes
    setVirtualBalance(selectedPlan.virtual_balance);
  }, [selectedPlan]);

  const totalPortfolioValue = portfolio.reduce((acc, p) => acc + p.current_value, 0);
  const totalProfitLoss = portfolio.reduce((acc, p) => acc + p.profit_loss, 0);
  const totalProfitLossPct = ((totalPortfolioValue + virtualBalance - selectedPlan.virtual_balance) / selectedPlan.virtual_balance) * 100;

  const validateForm = (): boolean => {
    const errors: Record<string, string> = {};
    if (!form.symbol.trim()) errors.symbol = 'Symbol is required';
    if (form.quantity <= 0) errors.quantity = 'Quantity must be greater than 0';
    if (form.price <= 0) errors.price = 'Price must be greater than 0';
    
    const totalValue = form.quantity * form.price;
    if (form.side === 'BUY' && totalValue > virtualBalance) {
      errors.quantity = `Insufficient balance. Max: ${(virtualBalance / form.price).toFixed(6)}`;
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleCreateSignal = async () => {
    if (!validateForm()) return;
    try {
      await createSignal({ ...form, plan_id: selectedPlan.id });
      // Update virtual balance (simulate)
      if (form.side === 'BUY') {
        setVirtualBalance(prev => prev - (form.quantity * form.price));
      } else {
        setVirtualBalance(prev => prev + (form.quantity * form.price));
      }
      setIsCreateModalOpen(false);
      setForm(initialFormState);
    } catch (error) {
      console.error('Failed to create signal:', error);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Trader Dashboard</h1>
          <p className="text-gray-400 mt-1">Execute trades with virtual balance</p>
        </div>
        <div className="flex items-center gap-4">
          <Select
            value={selectedPlan.id.toString()}
            onChange={(e) => {
              const plan = mockPlans.find(p => p.id === parseInt(e.target.value));
              if (plan) setSelectedPlan(plan);
            }}
            options={mockPlans.map(p => ({ 
              value: p.id.toString(), 
              label: `${p.name} - $${p.virtual_balance.toLocaleString()} Virtual` 
            }))}
          />
        </div>
      </div>

      {/* Virtual Balance Card */}
      <div className={`rounded-2xl p-6 ${
        isDark 
          ? 'bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600' 
          : 'bg-gradient-to-br from-blue-500 via-indigo-500 to-purple-500'
      }`}>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-white/80 text-sm font-medium">Virtual Balance</p>
            <p className="text-4xl font-bold text-white mt-1">
              {formatCurrency(virtualBalance)}
            </p>
            <p className="text-white/60 text-sm mt-2">
              Plan: {selectedPlan.name} • Min Investment: ${selectedPlan.min_investment.toLocaleString()}
            </p>
          </div>
          <div className="text-white/80">
            <FaWallet className="text-6xl opacity-50" />
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Portfolio Value"
          value={formatCurrency(totalPortfolioValue + virtualBalance)}
          icon={<FaWallet className="text-blue-500" />}
        />
        <StatCard
          title="Total P/L"
          value={formatCurrency(totalProfitLoss)}
          change={`${totalProfitLossPct >= 0 ? '+' : ''}${totalProfitLossPct.toFixed(2)}%`}
          changeType={totalProfitLoss >= 0 ? 'positive' : 'negative'}
          icon={<FaChartLine className={totalProfitLoss >= 0 ? 'text-green-500' : 'text-red-500'} />}
        />
        <StatCard
          title="Open Positions"
          value={portfolio.length.toString()}
          icon={<FaExchangeAlt className="text-purple-500" />}
        />
        <StatCard
          title="Total Signals"
          value={signals.length.toString()}
          icon={<FaTrophy className="text-yellow-500" />}
        />
      </div>

      {/* Virtual Portfolio */}
      <Card>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <FaCoins className="text-yellow-500" />
          Virtual Portfolio
        </h3>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="border-b border-gray-700">
                <th className="text-left py-3 px-4 font-medium text-gray-400">Asset</th>
                <th className="text-right py-3 px-4 font-medium text-gray-400">Quantity</th>
                <th className="text-right py-3 px-4 font-medium text-gray-400">Avg Price</th>
                <th className="text-right py-3 px-4 font-medium text-gray-400">Value</th>
                <th className="text-right py-3 px-4 font-medium text-gray-400">P/L</th>
                <th className="text-right py-3 px-4 font-medium text-gray-400">P/L %</th>
              </tr>
            </thead>
            <tbody>
              {portfolio.map((position) => (
                <tr key={position.symbol} className="border-b border-gray-800 hover:bg-gray-800/50">
                  <td className="py-3 px-4 font-medium text-white">{position.symbol}</td>
                  <td className="py-3 px-4 text-right text-gray-300">{formatNumber(position.quantity, 4)}</td>
                  <td className="py-3 px-4 text-right text-gray-300">{formatCurrency(position.avg_buy_price)}</td>
                  <td className="py-3 px-4 text-right text-white font-medium">{formatCurrency(position.current_value)}</td>
                  <td className={`py-3 px-4 text-right font-medium ${position.profit_loss >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                    {position.profit_loss >= 0 ? '+' : ''}{formatCurrency(position.profit_loss)}
                  </td>
                  <td className={`py-3 px-4 text-right font-medium ${position.profit_loss_pct >= 0 ? 'text-green-400' : 'text-red-400'}`}>
                    {position.profit_loss_pct >= 0 ? '+' : ''}{position.profit_loss_pct.toFixed(2)}%
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Trading Section */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <FaExchangeAlt className="text-blue-500" />
            Execute Trade
          </h3>
          <Button onClick={() => setIsCreateModalOpen(true)} size="sm">
            New Trade
          </Button>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Buy Section */}
          <div className={`p-4 rounded-xl ${isDark ? 'bg-green-900/20 border border-green-800' : 'bg-green-50 border border-green-200'}`}>
            <h4 className="text-green-400 font-semibold mb-3">Buy</h4>
            <div className="space-y-2">
              <Input 
                placeholder="Symbol (e.g., BTC)" 
                className="bg-transparent border-gray-700 text-white"
              />
              <div className="grid grid-cols-2 gap-2">
                <Input 
                  placeholder="Price" 
                  type="number"
                  className="bg-transparent border-gray-700 text-white"
                />
                <Input 
                  placeholder="Quantity" 
                  type="number"
                  className="bg-transparent border-gray-700 text-white"
                />
              </div>
              <Button className="w-full bg-green-600 hover:bg-green-700">
                Buy
              </Button>
            </div>
          </div>

          {/* Sell Section */}
          <div className={`p-4 rounded-xl ${isDark ? 'bg-red-900/20 border border-red-800' : 'bg-red-50 border border-red-200'}`}>
            <h4 className="text-red-400 font-semibold mb-3">Sell</h4>
            <div className="space-y-2">
              <Input 
                placeholder="Symbol (e.g., BTC)" 
                className="bg-transparent border-gray-700 text-white"
              />
              <div className="grid grid-cols-2 gap-2">
                <Input 
                  placeholder="Price" 
                  type="number"
                  className="bg-transparent border-gray-700 text-white"
                />
                <Input 
                  placeholder="Quantity" 
                  type="number"
                  className="bg-transparent border-gray-700 text-white"
                />
              </div>
              <Button className="w-full bg-red-600 hover:bg-red-700">
                Sell
              </Button>
            </div>
          </div>
        </div>
      </Card>

      {/* Signals List */}
      <Card>
        <h3 className="text-lg font-semibold text-white mb-4 flex items-center gap-2">
          <FaTrophy className="text-yellow-500" />
          Recent Signals
        </h3>
        {signalsLoading ? (
          <div className="animate-pulse space-y-3">
            {[...Array(3)].map((_, i) => <div key={i} className="h-12 bg-gray-800 rounded" />)}
          </div>
        ) : signals.length === 0 ? (
          <p className="text-gray-400 text-center py-8">No signals created yet.</p>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-gray-700">
                  <th className="text-left py-3 px-4 font-medium text-gray-400">Symbol</th>
                  <th className="text-left py-3 px-4 font-medium text-gray-400">Side</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-400">Price</th>
                  <th className="text-right py-3 px-4 font-medium text-gray-400">Quantity</th>
                  <th className="text-left py-3 px-4 font-medium text-gray-400">Plan</th>
                </tr>
              </thead>
              <tbody>
                {signals.slice(0, 5).map((signal) => (
                  <tr key={signal.signal_id} className="border-b border-gray-800 hover:bg-gray-800/50">
                    <td className="py-3 px-4 font-medium text-white">{signal.symbol}</td>
                    <td className="py-3 px-4">
                      <span className={`px-2 py-1 rounded text-xs font-medium ${
                        signal.side === 'BUY' 
                          ? 'bg-green-900/50 text-green-400' 
                          : 'bg-red-900/50 text-red-400'
                      }`}>
                        {signal.side}
                      </span>
                    </td>
                    <td className="py-3 px-4 text-right text-gray-300">{formatCurrency(signal.price)}</td>
                    <td className="py-3 px-4 text-right text-gray-300">{formatNumber(signal.quantity, 4)}</td>
                    <td className="py-3 px-4 text-gray-400">
                      <span className="flex items-center gap-1">
                        <FaCrown className="text-yellow-500 text-xs" />
                        Basic
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      {/* Create Signal Modal */}
      <Modal isOpen={isCreateModalOpen} onClose={() => setIsCreateModalOpen(false)} title="Execute Trade" size="lg">
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Input 
              label="Symbol" 
              placeholder="BTCUSDT" 
              value={form.symbol} 
              onChange={(e) => setForm({ ...form, symbol: e.target.value.toUpperCase() })} 
              error={formErrors.symbol} 
            />
            <Select 
              label="Side" 
              options={[{ value: 'BUY', label: 'BUY' }, { value: 'SELL', label: 'SELL' }]} 
              value={form.side} 
              onChange={(e) => setForm({ ...form, side: e.target.value as 'BUY' | 'SELL' })} 
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <Input 
              label="Price" 
              type="number" 
              step="0.01" 
              value={form.price || ''} 
              onChange={(e) => setForm({ ...form, price: parseFloat(e.target.value) || 0 })} 
              error={formErrors.price} 
            />
            <Input 
              label="Quantity" 
              type="number" 
              step="0.0001" 
              value={form.quantity || ''} 
              onChange={(e) => setForm({ ...form, quantity: parseFloat(e.target.value) || 0 })} 
              error={formErrors.quantity} 
            />
          </div>
          <div className={`p-4 rounded-lg ${isDark ? 'bg-gray-800' : 'bg-gray-100'}`}>
            <p className="text-gray-400 text-sm">Total Value</p>
            <p className="text-2xl font-bold text-white">
              {formatCurrency(form.price * form.quantity)}
            </p>
            <p className="text-gray-500 text-sm mt-1">
              Available: {formatCurrency(virtualBalance)}
            </p>
          </div>
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setIsCreateModalOpen(false)}>Cancel</Button>
            <Button onClick={handleCreateSignal} isLoading={signalsLoading}>
              Execute Trade
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}
