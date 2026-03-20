'use client';

import React from 'react';
import Navbar from '@/components/layout/Navbar';
import Sidebar from '@/components/layout/Sidebar';
import Footer from '@/components/layout/Footer';
import ErrorBoundary from '@/components/layout/ErrorBoundary';

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-gray-50 dark:bg-gray-950">
        <Navbar />
        <div className="flex">
          <Sidebar />
          <main className="flex-1 p-4 lg:p-6 min-h-[calc(100vh-4rem)]">
            {children}
          </main>
        </div>
        <Footer />
      </div>
    </ErrorBoundary>
  );
}
