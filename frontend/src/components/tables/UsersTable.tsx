'use client';

import React from 'react';
import type { User } from '@/types';
import Badge from '@/components/ui/Badge';
import { formatDate } from '@/utils/helpers';

interface UsersTableProps {
  users: User[];
  onEdit?: (user: User) => void;
  onSuspend?: (userId: number) => void;
  onDelete?: (userId: number) => void;
}

export default function UsersTable({ users, onEdit, onSuspend, onDelete }: UsersTableProps) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-sm">
        <thead>
          <tr className="border-b border-gray-200 dark:border-gray-700">
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">User</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Email</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Role</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Plan</th>
            <th className="text-left py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Joined</th>
            <th className="text-right py-3 px-4 font-medium text-gray-500 dark:text-gray-400">Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id} className="border-b border-gray-100 dark:border-gray-800 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors">
              <td className="py-3 px-4">
                <div className="flex items-center gap-3">
                  <div className="w-8 h-8 bg-blue-600 rounded-full flex items-center justify-center text-white text-sm font-medium">
                    {user.name.charAt(0).toUpperCase()}
                  </div>
                  <span className="font-medium text-gray-900 dark:text-white">{user.name}</span>
                </div>
              </td>
              <td className="py-3 px-4 text-gray-600 dark:text-gray-400">{user.email}</td>
              <td className="py-3 px-4">
                <Badge variant={user.role === 'admin' ? 'danger' : user.role === 'trader' ? 'warning' : 'info'}>
                  {user.role}
                </Badge>
              </td>
              <td className="py-3 px-4 text-gray-600 dark:text-gray-400 capitalize">
                {user.plan?.name || 'N/A'}
              </td>
              <td className="py-3 px-4 text-gray-500 dark:text-gray-400 text-xs">
                {formatDate(user.created_at)}
              </td>
              <td className="py-3 px-4 text-right">
                <div className="flex items-center justify-end gap-2">
                  {onEdit && (
                    <button onClick={() => onEdit(user)} className="text-blue-500 hover:text-blue-700 text-xs font-medium">
                      Edit
                    </button>
                  )}
                  {onSuspend && (
                    <button onClick={() => onSuspend(user.id)} className="text-yellow-500 hover:text-yellow-700 text-xs font-medium">
                      Suspend
                    </button>
                  )}
                  {onDelete && (
                    <button onClick={() => onDelete(user.id)} className="text-red-500 hover:text-red-700 text-xs font-medium">
                      Delete
                    </button>
                  )}
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
