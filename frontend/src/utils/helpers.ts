import type { OrderStatus } from '@/types';

export function formatCurrency(value: number, currency = 'USD'): string {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency }).format(value);
}

export function formatNumber(value: number, decimals = 2): string {
  return new Intl.NumberFormat('en-US', { minimumFractionDigits: decimals, maximumFractionDigits: decimals }).format(value);
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
  });
}

export function formatDateTime(dateStr: string): string {
  return new Date(dateStr).toLocaleString('en-US', {
    year: 'numeric', month: 'short', day: 'numeric',
    hour: '2-digit', minute: '2-digit', second: '2-digit',
  });
}

export function formatPercentage(value: number): string {
  return `${value >= 0 ? '+' : ''}${value.toFixed(2)}%`;
}

export function getStatusColor(status: OrderStatus): string {
  switch (status) {
    case 'FILLED': return 'text-green-500';
    case 'FAILED': return 'text-red-500';
    case 'CANCELED': return 'text-gray-500';
    case 'EXECUTING': return 'text-yellow-500';
    case 'PENDING': return 'text-blue-500';
    default: return 'text-gray-400';
  }
}

export function getStatusBgColor(status: OrderStatus): string {
  switch (status) {
    case 'FILLED': return 'bg-green-500/10 text-green-500 border-green-500/20';
    case 'FAILED': return 'bg-red-500/10 text-red-500 border-red-500/20';
    case 'CANCELED': return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
    case 'EXECUTING': return 'bg-yellow-500/10 text-yellow-500 border-yellow-500/20';
    case 'PENDING': return 'bg-blue-500/10 text-blue-500 border-blue-500/20';
    default: return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
  }
}

export function cn(...classes: (string | undefined | null | false)[]): string {
  return classes.filter(Boolean).join(' ');
}

export function truncate(str: string, length: number): string {
  if (str.length <= length) return str;
  return str.slice(0, length) + '...';
}

export function validateEmail(email: string): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
}

export function validatePassword(password: string): string[] {
  const errors: string[] = [];
  if (password.length < 8) errors.push('Password must be at least 8 characters');
  if (!/[A-Z]/.test(password)) errors.push('Must contain an uppercase letter');
  if (!/[a-z]/.test(password)) errors.push('Must contain a lowercase letter');
  if (!/[0-9]/.test(password)) errors.push('Must contain a number');
  return errors;
}
