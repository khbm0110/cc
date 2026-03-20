// ===== User & Auth Types =====
export type UserRole = 'client' | 'trader' | 'admin';

export interface User {
  id: number;
  name: string;
  email: string;
  role: UserRole;
  plan_id: number;
  plan?: Plan;
  created_at: string;
  updated_at: string;
}

export interface Plan {
  id: number;
  name: string;
  max_exposure_ratio: number;
  order_limit_per_min: number;
  created_at: string;
  updated_at: string;
}

export interface AuthTokens {
  access_token: string;
  refresh_token: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  name: string;
  email: string;
  password: string;
  plan_id: number;
  role: UserRole;
}

// ===== Order Types =====
export type OrderStatus = 'PENDING' | 'EXECUTING' | 'FILLED' | 'FAILED' | 'CANCELED';

export interface Order {
  id: number;
  user_id: number;
  client_order_id: string;
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  status: OrderStatus;
  binance_order_id?: number;
  error_message?: string;
  retry_count: number;
  created_at: string;
  updated_at: string;
}

export interface OrderFilters {
  status?: OrderStatus;
  symbol?: string;
  side?: 'BUY' | 'SELL';
  from_date?: string;
  to_date?: string;
  page: number;
  limit: number;
}

// ===== Trade Signal Types =====
export interface TradeSignal {
  signal_id: string;
  user_id: number;
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  stop_loss?: number;
  take_profit?: number;
  max_slippage?: number;
  client_order_id: string;
  created_at?: string;
}

export interface CreateSignalRequest {
  symbol: string;
  side: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  stop_loss?: number;
  take_profit?: number;
  max_slippage?: number;
}

// ===== Metrics Types =====
export interface SystemMetrics {
  total_orders: number;
  filled_orders: number;
  failed_orders: number;
  pending_orders: number;
  rate_limit_hits: number;
  active_users: number;
  circuit_breaker_open: number;
}

export interface ReconcilerLog {
  id: number;
  order_id: number;
  action: string;
  previous_status: OrderStatus;
  new_status: OrderStatus;
  details: string;
  created_at: string;
}

// ===== Dashboard Types =====
export interface PortfolioSummary {
  total_balance: number;
  available_balance: number;
  total_exposure: number;
  exposure_ratio: number;
  total_pnl: number;
  pnl_percentage: number;
  total_trades: number;
  winning_trades: number;
}

export interface ChartDataPoint {
  date: string;
  value: number;
  label?: string;
}

export interface PerformanceMetrics {
  win_rate: number;
  avg_profit: number;
  avg_loss: number;
  total_signals: number;
  successful_signals: number;
  profit_factor: number;
  sharpe_ratio: number;
}

// ===== API Response =====
export interface ApiResponse<T> {
  data: T;
  message?: string;
  success: boolean;
}

export interface PaginatedResponse<T> {
  data: T[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
}

// ===== Notification Types =====
export interface NotificationPreferences {
  email_on_trade: boolean;
  email_on_error: boolean;
  push_notifications: boolean;
  trade_alerts: boolean;
}
