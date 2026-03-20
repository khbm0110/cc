'use client';

import React from 'react';
import type { Order } from '@/types';
import { StatusBadge } from '@/components/ui/Badge';
import { formatDateTime, formatCurrency, formatNumber } from '@/utils/helpers';
import EmptyState from '@/components/ui/EmptyState';

interface OrdersTableProps {
  orders: Order[];
  onCancel?: (id: number) => void;
  isLoading?: boolean;
}

export default function OrdersTable({ orders, onCancel, isLoading }: OrdersTableProps) {
  if (isLoading) {
    return (
      <div className="animate-pulse space-y-3">
        {[...Array(5)].map((_, i) => (
          <div key={i} className="h-12 bg-gray-200 dark:bg-gray-700 rounded" />
        ))}
      </div>
    );
  }

  if (orders.length === 0) {
    return (
      <EmptyState
        title="No orders found"
        description="Orders will appear here once trades are executed."
        icon={
          <svg className="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
        }
      />
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">ID</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Symbol</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Side</th>
            <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Price</th>
            <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Quantity</th>
            <th className="text-center py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Status</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Time</th>
            <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Actions</th>
          </tr>
        </thead>
        <tbody>
          {orders.map((order) => (
            <tr key={order.id} className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
              <td className="py-3 px-4 font-mono text-xs text-gray-600 dark:text-gray-400">
                #{order.id}
              </td>
              <td className="py-3 px-4 font-medium text-gray-900 dark:text-white">
                {order.symbol}
              </td>
              <td className="py-3 px-4">
                <span className={order.side === 'BUY' ? 'text-green-500 font-medium' : 'text-red-500 font-medium'}>
                  {order.side}
                </span>
              </td>
              <td className="py-3 px-4 text-right font-mono text-gray-900 dark:text-white">
                {formatCurrency(order.price)}
              </td>
              <td className="py-3 px-4 text-right font-mono text-gray-600 dark:text-gray-400">
                {formatNumber(order.quantity, 4)}
              </td>
              <td className="py-3 px-4 text-center">
                <StatusBadge status={order.status} />
              </td>
              <td className="py-3 px-4 text-gray-500 dark:text-gray-400 text-xs">
                {formatDateTime(order.created_at)}
              </td>
              <td className="py-3 px-4 text-right">
                {order.status === 'PENDING' && onCancel && (
                  <button
                    onClick={() => onCancel(order.id)}
                    className="text-red-500 hover:text-red-700 text-xs font-medium transition-colors"
                  >
                    Cancel
                  </button>
                )}
                {order.error_message && (
                  <span className="text-red-400 text-xs" title={order.error_message}>
                    Error
                  </span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
