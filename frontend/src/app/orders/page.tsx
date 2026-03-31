'use client';

import React, { useEffect, useState } from 'react';
import Card from '@/components/ui/Card';
import Select from '@/components/ui/Select';
import Input from '@/components/ui/Input';
import OrdersTable from '@/components/tables/OrdersTable';
import Pagination from '@/components/tables/Pagination';
import { useOrderStore } from '@/features/orderStore';
import type { OrderStatus } from '@/types';

export default function OrdersPage() {
  const { orders, totalOrders, totalPages, currentPage, isLoading, fetchOrders, cancelOrder, setFilters, filters } = useOrderStore();
  const [search, setSearch] = useState('');

  useEffect(() => {
    fetchOrders();
  }, [fetchOrders]);

  const handleStatusFilter = (status: string) => {
    setFilters({ status: (status || undefined) as OrderStatus | undefined });
    fetchOrders({ status: (status || undefined) as OrderStatus | undefined, page: 1 });
  };

  const handleSideFilter = (side: string) => {
    setFilters({ side: (side || undefined) as 'BUY' | 'SELL' | undefined });
    fetchOrders({ side: (side || undefined) as 'BUY' | 'SELL' | undefined, page: 1 });
  };

  const handleSearch = (value: string) => {
    setSearch(value);
    setFilters({ symbol: value || undefined });
    fetchOrders({ symbol: value || undefined, page: 1 });
  };

  const handlePageChange = (page: number) => {
    fetchOrders({ page });
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Order History</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">
          {totalOrders} total orders
        </p>
      </div>

      {/* Filters */}
      <Card>
        <div className="flex flex-col sm:flex-row gap-4">
          <div className="flex-1">
            <Input
              placeholder="Search by symbol..."
              value={search}
              onChange={(e) => handleSearch(e.target.value)}
              icon={
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                </svg>
              }
            />
          </div>
          <div className="w-full sm:w-40">
            <Select
              options={[
                { value: '', label: 'All Statuses' },
                { value: 'PENDING', label: 'Pending' },
                { value: 'EXECUTING', label: 'Executing' },
                { value: 'FILLED', label: 'Filled' },
                { value: 'FAILED', label: 'Failed' },
                { value: 'CANCELED', label: 'Canceled' },
              ]}
              value={filters.status || ''}
              onChange={(e) => handleStatusFilter(e.target.value)}
            />
          </div>
          <div className="w-full sm:w-32">
            <Select
              options={[
                { value: '', label: 'All Sides' },
                { value: 'BUY', label: 'Buy' },
                { value: 'SELL', label: 'Sell' },
              ]}
              value={filters.side || ''}
              onChange={(e) => handleSideFilter(e.target.value)}
            />
          </div>
          <div className="w-full sm:w-40">
            <Input
              type="date"
              value={filters.from_date || ''}
              onChange={(e) => {
                setFilters({ from_date: e.target.value || undefined });
                fetchOrders({ from_date: e.target.value || undefined, page: 1 });
              }}
            />
          </div>
        </div>
      </Card>

      {/* Orders Table */}
      <Card padding="none">
        <div className="p-6">
          <OrdersTable orders={orders} onCancel={cancelOrder} isLoading={isLoading} />
          <Pagination currentPage={currentPage} totalPages={totalPages} onPageChange={handlePageChange} />
        </div>
      </Card>
    </div>
  );
}
