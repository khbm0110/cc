'use client';

import React, { useEffect, useState } from 'react';
import Card, { StatCard } from '@/components/ui/Card';
import OrdersTable from '@/components/tables/OrdersTable';
import PortfolioChart from '@/components/charts/PortfolioChart';
import OrderStatusChart from '@/components/charts/OrderStatusChart';
import { useOrderStore } from '@/features/orderStore';
import { useAuthStore } from '@/features/authStore';
import { formatCurrency, formatPercentage } from '@/utils/helpers';
import type { ChartDataPoint, PortfolioSummary } from '@/types';

// Mock data for demo
const mockPortfolio: PortfolioSummary = {
  total_balance: 25430.50,
  available_balance: 18200.75,
  total_exposure: 7229.75,
  exposure_ratio: 0.284,
  total_pnl: 3245.80,
  pnl_percentage: 14.62,
  total_trades: 156,
  winning_trades: 112,
};

const mockChartData: ChartDataPoint[] = [
  { date: 'Jan', value: 20000 },
  { date: 'Feb', value: 21200 },
  { date: 'Mar', value: 20800 },
  { date: 'Apr', value: 22500 },
  { date: 'May', value: 23100 },
  { date: 'Jun', value: 24200 },
  { date: 'Jul', value: 23800 },
  { date: 'Aug', value: 25430 },
];

export default function ClientDashboard() {
  const { orders, fetchOrders, cancelOrder, isLoading } = useOrderStore();
  const { user } = useAuthStore();
  const [portfolio] = useState<PortfolioSummary>(mockPortfolio);

  useEffect(() => {
    fetchOrders({ limit: 10 });
  }, [fetchOrders]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">
          Welcome back{user?.name ? `, ${user.name}` : ''}
        </h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">Here&apos;s your portfolio overview</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard
          title="Total Balance"
          value={formatCurrency(portfolio.total_balance)}
          change={formatPercentage(portfolio.pnl_percentage)}
          changeType="positive"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          }
        />
        <StatCard
          title="Available Balance"
          value={formatCurrency(portfolio.available_balance)}
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
          }
        />
        <StatCard
          title="Total P&L"
          value={formatCurrency(portfolio.total_pnl)}
          change={formatPercentage(portfolio.pnl_percentage)}
          changeType="positive"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
            </svg>
          }
        />
        <StatCard
          title="Win Rate"
          value={`${((portfolio.winning_trades / portfolio.total_trades) * 100).toFixed(1)}%`}
          change={`${portfolio.winning_trades}/${portfolio.total_trades} trades`}
          changeType="neutral"
          icon={
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          }
        />
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <Card className="lg:col-span-2">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Portfolio Value</h3>
          <PortfolioChart data={mockChartData} />
        </Card>
        <Card>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Order Distribution</h3>
          <OrderStatusChart filled={112} failed={18} pending={14} canceled={12} />
        </Card>
      </div>

      {/* Plan Info */}
      {user?.plan && (
        <Card>
          <div className="flex items-center justify-between">
            <div>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Subscription Plan</h3>
              <p className="text-gray-500 dark:text-gray-400 mt-1">
                <span className="capitalize font-medium text-blue-600 dark:text-blue-400">{user.plan.name}</span>
                {' '}&mdash; {user.plan.order_limit_per_min} orders/min, {(user.plan.max_exposure_ratio * 100).toFixed(0)}% max exposure
              </p>
            </div>
            <div className="text-right">
              <p className="text-sm text-gray-500 dark:text-gray-400">Exposure Usage</p>
              <div className="w-48 h-2 bg-gray-200 dark:bg-gray-700 rounded-full mt-2">
                <div
                  className="h-full bg-blue-600 rounded-full"
                  style={{ width: `${Math.min(portfolio.exposure_ratio / user.plan.max_exposure_ratio * 100, 100)}%` }}
                />
              </div>
            </div>
          </div>
        </Card>
      )}

      {/* Recent Orders */}
      <Card>
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Recent Trades</h3>
          <a href="/orders" className="text-sm text-blue-600 hover:text-blue-700 dark:text-blue-400 font-medium">
            View all
          </a>
        </div>
        <OrdersTable orders={orders} onCancel={cancelOrder} isLoading={isLoading} />
      </Card>
    </div>
  );
}
