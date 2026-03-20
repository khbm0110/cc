'use client';

import React, { useEffect, useState, useCallback } from 'react';
import Card, { StatCard } from '@/components/ui/Card';
import UsersTable from '@/components/tables/UsersTable';
import OrderStatusChart from '@/components/charts/OrderStatusChart';
import Modal from '@/components/ui/Modal';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import ConfirmDialog from '@/components/ui/ConfirmDialog';
import Pagination from '@/components/tables/Pagination';
import Spinner from '@/components/ui/Spinner';
import { adminService } from '@/services/admin';
import { formatDateTime } from '@/utils/helpers';
import type { User, Plan, SystemMetrics, ReconcilerLog } from '@/types';

export default function AdminDashboard() {
  const [users, setUsers] = useState<User[]>([]);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [reconcilerLogs, setReconcilerLogs] = useState<ReconcilerLog[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [usersPage, setUsersPage] = useState(1);
  const [usersTotalPages, setUsersTotalPages] = useState(1);
  const [activeTab, setActiveTab] = useState<'overview' | 'users' | 'plans' | 'reconciler'>('overview');

  // Plan modal state
  const [isPlanModalOpen, setIsPlanModalOpen] = useState(false);
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null);
  const [planForm, setPlanForm] = useState({ name: '', max_exposure_ratio: 0.5, order_limit_per_min: 10 });

  // Confirm dialog
  const [confirmAction, setConfirmAction] = useState<{ type: string; id: number; message: string } | null>(null);

  const loadData = useCallback(async () => {
    setIsLoading(true);
    try {
      const [usersData, plansData, metricsData, logsData] = await Promise.allSettled([
        adminService.getUsers(usersPage),
        adminService.getPlans(),
        adminService.getSystemMetrics(),
        adminService.getReconcilerLogs(),
      ]);
      if (usersData.status === 'fulfilled') {
        setUsers(usersData.value.data);
        setUsersTotalPages(usersData.value.total_pages);
      }
      if (plansData.status === 'fulfilled') setPlans(plansData.value);
      if (metricsData.status === 'fulfilled') setMetrics(metricsData.value);
      if (logsData.status === 'fulfilled') setReconcilerLogs(logsData.value.data);
    } finally {
      setIsLoading(false);
    }
  }, [usersPage]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSuspendUser = async (userId: number) => {
    try {
      await adminService.suspendUser(userId);
      await loadData();
    } catch { /* error */ }
  };

  const handleDeleteUser = async (userId: number) => {
    try {
      await adminService.deleteUser(userId);
      await loadData();
    } catch { /* error */ }
  };

  const handleSavePlan = async () => {
    try {
      if (editingPlan) {
        await adminService.updatePlan(editingPlan.id, planForm);
      } else {
        await adminService.createPlan(planForm);
      }
      setIsPlanModalOpen(false);
      setEditingPlan(null);
      setPlanForm({ name: '', max_exposure_ratio: 0.5, order_limit_per_min: 10 });
      await loadData();
    } catch { /* error */ }
  };

  const handleDeletePlan = async (planId: number) => {
    try {
      await adminService.deletePlan(planId);
      await loadData();
    } catch { /* error */ }
  };

  const tabs = [
    { key: 'overview', label: 'Overview' },
    { key: 'users', label: 'Users' },
    { key: 'plans', label: 'Plans' },
    { key: 'reconciler', label: 'Reconciler' },
  ] as const;

  if (isLoading && !metrics) {
    return <div className="flex items-center justify-center min-h-[400px]"><Spinner size="lg" /></div>;
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Admin Dashboard</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">System overview and management</p>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <div className="flex gap-4">
          {tabs.map((tab) => (
            <button
              key={tab.key}
              onClick={() => setActiveTab(tab.key)}
              className={`pb-3 px-1 text-sm font-medium border-b-2 transition-colors ${
                activeTab === tab.key
                  ? 'border-blue-600 text-blue-600 dark:text-blue-400'
                  : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      </div>

      {/* Overview Tab */}
      {activeTab === 'overview' && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <StatCard title="Total Orders" value={metrics?.total_orders ?? 0} icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" /></svg>} />
            <StatCard title="Active Users" value={metrics?.active_users ?? 0} icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" /></svg>} />
            <StatCard title="Rate Limit Hits" value={metrics?.rate_limit_hits ?? 0} changeType="negative" icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L4.082 16.5c-.77.833.192 2.5 1.732 2.5z" /></svg>} />
            <StatCard title="Circuit Breakers Open" value={metrics?.circuit_breaker_open ?? 0} changeType={metrics?.circuit_breaker_open ? 'negative' : 'neutral'} icon={<svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" /></svg>} />
          </div>
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            <Card>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Order Status Distribution</h3>
              <OrderStatusChart
                filled={metrics?.filled_orders ?? 0}
                failed={metrics?.failed_orders ?? 0}
                pending={metrics?.pending_orders ?? 0}
                canceled={0}
              />
            </Card>
            <Card>
              <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">System Health</h3>
              <div className="space-y-4">
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Worker Service</span>
                  <span className="flex items-center gap-2 text-sm font-medium text-green-500"><span className="w-2 h-2 bg-green-500 rounded-full" /> Healthy</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Reconciler Service</span>
                  <span className="flex items-center gap-2 text-sm font-medium text-green-500"><span className="w-2 h-2 bg-green-500 rounded-full" /> Healthy</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <span className="text-sm text-gray-600 dark:text-gray-400">Redis Event Bus</span>
                  <span className="flex items-center gap-2 text-sm font-medium text-green-500"><span className="w-2 h-2 bg-green-500 rounded-full" /> Connected</span>
                </div>
                <div className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                  <span className="text-sm text-gray-600 dark:text-gray-400">PostgreSQL</span>
                  <span className="flex items-center gap-2 text-sm font-medium text-green-500"><span className="w-2 h-2 bg-green-500 rounded-full" /> Connected</span>
                </div>
              </div>
            </Card>
          </div>
        </>
      )}

      {/* Users Tab */}
      {activeTab === 'users' && (
        <Card>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">User Management</h3>
          </div>
          <UsersTable
            users={users}
            onSuspend={(id) => setConfirmAction({ type: 'suspend', id, message: 'Are you sure you want to suspend this user?' })}
            onDelete={(id) => setConfirmAction({ type: 'delete', id, message: 'This action cannot be undone. Are you sure?' })}
          />
          <Pagination currentPage={usersPage} totalPages={usersTotalPages} onPageChange={setUsersPage} />
        </Card>
      )}

      {/* Plans Tab */}
      {activeTab === 'plans' && (
        <Card>
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white">Plan Management</h3>
            <Button size="sm" onClick={() => { setEditingPlan(null); setPlanForm({ name: '', max_exposure_ratio: 0.5, order_limit_per_min: 10 }); setIsPlanModalOpen(true); }}>
              Add Plan
            </Button>
          </div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {plans.map((plan) => (
              <div key={plan.id} className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                <h4 className="font-semibold text-gray-900 dark:text-white capitalize">{plan.name}</h4>
                <div className="mt-2 space-y-1 text-sm text-gray-500 dark:text-gray-400">
                  <p>Max Exposure: {(plan.max_exposure_ratio * 100).toFixed(0)}%</p>
                  <p>Orders/min: {plan.order_limit_per_min}</p>
                </div>
                <div className="mt-3 flex gap-2">
                  <Button variant="ghost" size="sm" onClick={() => { setEditingPlan(plan); setPlanForm({ name: plan.name, max_exposure_ratio: plan.max_exposure_ratio, order_limit_per_min: plan.order_limit_per_min }); setIsPlanModalOpen(true); }}>Edit</Button>
                  <Button variant="ghost" size="sm" onClick={() => handleDeletePlan(plan.id)} className="text-red-500">Delete</Button>
                </div>
              </div>
            ))}
          </div>
        </Card>
      )}

      {/* Reconciler Tab */}
      {activeTab === 'reconciler' && (
        <Card>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Reconciler Logs</h3>
          {reconcilerLogs.length === 0 ? (
            <p className="text-gray-500 dark:text-gray-400 text-center py-8">No reconciler logs available</p>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-gray-200 dark:border-gray-700">
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Order ID</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Action</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">From</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">To</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Details</th>
                    <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Time</th>
                  </tr>
                </thead>
                <tbody>
                  {reconcilerLogs.map((log) => (
                    <tr key={log.id} className="border-b border-gray-100 dark:border-gray-800">
                      <td className="py-3 px-4 font-mono text-xs">#{log.order_id}</td>
                      <td className="py-3 px-4">{log.action}</td>
                      <td className="py-3 px-4">{log.previous_status}</td>
                      <td className="py-3 px-4">{log.new_status}</td>
                      <td className="py-3 px-4 text-gray-500 text-xs max-w-xs truncate">{log.details}</td>
                      <td className="py-3 px-4 text-xs text-gray-500">{formatDateTime(log.created_at)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </Card>
      )}

      {/* Plan Modal */}
      <Modal isOpen={isPlanModalOpen} onClose={() => setIsPlanModalOpen(false)} title={editingPlan ? 'Edit Plan' : 'Create Plan'}>
        <div className="space-y-4">
          <Input label="Plan Name" value={planForm.name} onChange={(e) => setPlanForm({ ...planForm, name: e.target.value })} />
          <Input label="Max Exposure Ratio" type="number" step="0.01" value={planForm.max_exposure_ratio} onChange={(e) => setPlanForm({ ...planForm, max_exposure_ratio: parseFloat(e.target.value) })} />
          <Input label="Orders per Minute" type="number" value={planForm.order_limit_per_min} onChange={(e) => setPlanForm({ ...planForm, order_limit_per_min: parseInt(e.target.value) })} />
          <div className="flex justify-end gap-3 pt-4">
            <Button variant="ghost" onClick={() => setIsPlanModalOpen(false)}>Cancel</Button>
            <Button onClick={handleSavePlan}>{editingPlan ? 'Update' : 'Create'}</Button>
          </div>
        </div>
      </Modal>

      {/* Confirm Dialog */}
      <ConfirmDialog
        isOpen={!!confirmAction}
        onClose={() => setConfirmAction(null)}
        onConfirm={async () => {
          if (confirmAction?.type === 'suspend') await handleSuspendUser(confirmAction.id);
          if (confirmAction?.type === 'delete') await handleDeleteUser(confirmAction.id);
          setConfirmAction(null);
        }}
        title={confirmAction?.type === 'suspend' ? 'Suspend User' : 'Delete User'}
        message={confirmAction?.message || ''}
      />
    </div>
  );
}
