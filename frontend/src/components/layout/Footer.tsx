'use client';

import React from 'react';

export default function Footer() {
  return (
    <footer className="border-t border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900 py-4 px-6">
      <div className="flex flex-col sm:flex-row items-center justify-between gap-2 text-sm text-gray-500 dark:text-gray-400">
        <p>&copy; {new Date().getFullYear()} CopyTrader. All rights reserved.</p>
        <div className="flex items-center gap-4">
          <a href="#" className="hover:text-gray-700 dark:hover:text-gray-300 transition-colors">Privacy</a>
          <a href="#" className="hover:text-gray-700 dark:hover:text-gray-300 transition-colors">Terms</a>
          <a href="#" className="hover:text-gray-700 dark:hover:text-gray-300 transition-colors">Support</a>
        </div>
      </div>
    </footer>
  );
}
