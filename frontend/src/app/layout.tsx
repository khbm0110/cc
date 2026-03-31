import type { Metadata } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { Toaster } from "react-hot-toast";
import "./globals.css";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

export const metadata: Metadata = {
  title: "CopyTrader - Binance Copy Trading Platform",
  description: "Professional copy trading platform for Binance. Automatically copy trades from expert traders with real-time execution, risk management, and portfolio tracking.",
  keywords: ["copy trading", "binance", "crypto", "trading platform", "automated trading"],
  openGraph: {
    title: "CopyTrader - Binance Copy Trading Platform",
    description: "Professional copy trading platform for Binance with real-time execution and risk management.",
    type: "website",
  },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${geistSans.variable} ${geistMono.variable} h-full antialiased dark`}
      suppressHydrationWarning
    >
      <body className="min-h-full flex flex-col">
        {children}
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 4000,
            style: {
              background: '#1F2937',
              color: '#F9FAFB',
              borderRadius: '8px',
              border: '1px solid #374151',
            },
            success: {
              iconTheme: { primary: '#10B981', secondary: '#F9FAFB' },
            },
            error: {
              iconTheme: { primary: '#EF4444', secondary: '#F9FAFB' },
            },
          }}
        />
      </body>
    </html>
  );
}
