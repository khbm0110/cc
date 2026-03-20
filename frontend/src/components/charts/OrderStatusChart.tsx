'use client';

import React from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Legend, Tooltip } from 'recharts';

interface OrderStatusChartProps {
  filled: number;
  failed: number;
  pending: number;
  canceled: number;
  height?: number;
}

const COLORS = ['#10B981', '#EF4444', '#3B82F6', '#6B7280'];

export default function OrderStatusChart({ filled, failed, pending, canceled, height = 250 }: OrderStatusChartProps) {
  const data = [
    { name: 'Filled', value: filled },
    { name: 'Failed', value: failed },
    { name: 'Pending', value: pending },
    { name: 'Canceled', value: canceled },
  ].filter((d) => d.value > 0);

  if (data.length === 0) {
    return (
      <div className="flex items-center justify-center" style={{ height }}>
        <p className="text-gray-500 dark:text-gray-400">No order data</p>
      </div>
    );
  }

  return (
    <ResponsiveContainer width="100%" height={height}>
      <PieChart>
        <Pie
          data={data}
          cx="50%"
          cy="50%"
          innerRadius={60}
          outerRadius={80}
          paddingAngle={5}
          dataKey="value"
        >
          {data.map((_, index) => (
            <Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />
          ))}
        </Pie>
        <Tooltip
          contentStyle={{
            backgroundColor: '#1F2937',
            border: '1px solid #374151',
            borderRadius: '8px',
            color: '#F9FAFB',
          }}
        />
        <Legend />
      </PieChart>
    </ResponsiveContainer>
  );
}
