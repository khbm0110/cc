'use client';

import React from 'react';
import { useAuthStore } from '@/features/authStore';
import ClientDashboard from './client/page';
import TraderDashboard from './trader/page';
import AdminDashboard from './admin/page';

export default function DashboardPage() {
  const { user } = useAuthStore();

  if (!user) {
    return <ClientDashboard />;
  }

  switch (user.role) {
    case 'admin':
      return <AdminDashboard />;
    case 'trader':
      return <TraderDashboard />;
    default:
      return <ClientDashboard />;
  }
}
