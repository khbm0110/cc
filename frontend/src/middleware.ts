import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

// Paths that don't require authentication
const publicPaths = [
  '/',
  '/auth/login',
  '/auth/register',
  '/auth/forgot-password',
  '/api/auth/login',
  '/api/auth/register',
  '/api/auth/refresh',
  '/api/auth/forgot-password',
];

// Paths that require admin role
const adminPaths = [
  '/dashboard/admin',
  '/api/admin',
];

// Paths that require trader role
const traderPaths = [
  '/dashboard/trader',
  '/api/signals',
  '/api/trades',
];

function isPublicPath(pathname: string): boolean {
  return publicPaths.some(path => 
    pathname === path || pathname.startsWith(path + '/')
  );
}

function isAdminPath(pathname: string): boolean {
  return adminPaths.some(path => 
    pathname.startsWith(path)
  );
}

function isTraderPath(pathname: string): boolean {
  return traderPaths.some(path => 
    pathname.startsWith(path)
  );
}

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public paths
  if (isPublicPath(pathname)) {
    return NextResponse.next();
  }

  // Check for token
  const token = request.cookies.get('access_token')?.value || 
                request.headers.get('authorization')?.replace('Bearer ', '');

  // If no token, redirect to login for protected paths
  if (!token) {
    const loginUrl = new URL('/auth/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  // For API routes, just verify token exists
  if (pathname.startsWith('/api/')) {
    // Add user info headers for downstream API handlers
    const response = NextResponse.next();
    
    // Optionally decode and pass user role
    try {
      // Simple JWT decode (in production, use proper JWT library)
      const base64Url = token.split('.')[1] || '{}';
      const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
      const jsonPayload = decodeURIComponent(
        atob(base64)
          .split('')
          .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
          .join('')
      );
      const payload = JSON.parse(jsonPayload);
      response.headers.set('x-user-id', String(payload.user_id || ''));
      response.headers.set('x-user-role', payload.role || '');
    } catch {
      // Invalid token format
    }
    
    return response;
  }

  // For page routes, check role-based access
  if (isAdminPath(pathname)) {
    // Check if user has admin role (would need to decode from token)
    // For now, we'll let the client-side handle this check
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization files)
     * - favicon.ico (favicon file)
     * - public files (public directory)
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
