'use client';

import React, { useState } from 'react';
import Card from '@/components/ui/Card';
import Input from '@/components/ui/Input';
import Button from '@/components/ui/Button';
import { useAuthStore } from '@/features/authStore';
import { authService } from '@/services/auth';
import { validatePassword } from '@/utils/helpers';
import toast from 'react-hot-toast';

export default function SettingsPage() {
  const { user } = useAuthStore();
  const [activeTab, setActiveTab] = useState<'profile' | 'security' | 'api-keys' | 'notifications'>('profile');

  // Profile form
  const [profileForm, setProfileForm] = useState({
    name: user?.name || '',
    email: user?.email || '',
  });
  const [profileLoading, setProfileLoading] = useState(false);

  // Password form
  const [passwordForm, setPasswordForm] = useState({
    oldPassword: '',
    newPassword: '',
    confirmPassword: '',
  });
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [passwordErrors, setPasswordErrors] = useState<Record<string, string>>({});

  // Notification preferences
  const [notifications, setNotifications] = useState({
    email_on_trade: true,
    email_on_error: true,
    push_notifications: false,
    trade_alerts: true,
  });

  const handleUpdateProfile = async (e: React.FormEvent) => {
    e.preventDefault();
    setProfileLoading(true);
    try {
      await authService.updateProfile(profileForm);
      toast.success('Profile updated successfully');
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to update profile');
    } finally {
      setProfileLoading(false);
    }
  };

  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    const errors: Record<string, string> = {};
    const pwErrors = validatePassword(passwordForm.newPassword);
    if (pwErrors.length > 0) errors.newPassword = pwErrors[0];
    if (passwordForm.newPassword !== passwordForm.confirmPassword) errors.confirmPassword = 'Passwords do not match';
    if (!passwordForm.oldPassword) errors.oldPassword = 'Current password is required';
    setPasswordErrors(errors);
    if (Object.keys(errors).length > 0) return;

    setPasswordLoading(true);
    try {
      await authService.changePassword(passwordForm.oldPassword, passwordForm.newPassword);
      toast.success('Password changed successfully');
      setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to change password');
    } finally {
      setPasswordLoading(false);
    }
  };

  const tabs = [
    { key: 'profile', label: 'Profile' },
    { key: 'security', label: 'Security' },
    { key: 'api-keys', label: 'API Keys' },
    { key: 'notifications', label: 'Notifications' },
  ] as const;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white">Settings</h1>
        <p className="text-gray-500 dark:text-gray-400 mt-1">Manage your account preferences</p>
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

      {/* Profile Tab */}
      {activeTab === 'profile' && (
        <Card>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Profile Information</h3>
          <form onSubmit={handleUpdateProfile} className="space-y-4 max-w-lg">
            <Input label="Full Name" value={profileForm.name} onChange={(e) => setProfileForm({ ...profileForm, name: e.target.value })} />
            <Input label="Email" type="email" value={profileForm.email} onChange={(e) => setProfileForm({ ...profileForm, email: e.target.value })} />
            <div className="pt-2">
              <Button type="submit" isLoading={profileLoading}>Save Changes</Button>
            </div>
          </form>
        </Card>
      )}

      {/* Security Tab */}
      {activeTab === 'security' && (
        <div className="space-y-6">
          <Card>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Change Password</h3>
            <form onSubmit={handleChangePassword} className="space-y-4 max-w-lg">
              <Input label="Current Password" type="password" value={passwordForm.oldPassword} onChange={(e) => setPasswordForm({ ...passwordForm, oldPassword: e.target.value })} error={passwordErrors.oldPassword} />
              <Input label="New Password" type="password" value={passwordForm.newPassword} onChange={(e) => setPasswordForm({ ...passwordForm, newPassword: e.target.value })} error={passwordErrors.newPassword} />
              <Input label="Confirm New Password" type="password" value={passwordForm.confirmPassword} onChange={(e) => setPasswordForm({ ...passwordForm, confirmPassword: e.target.value })} error={passwordErrors.confirmPassword} />
              <div className="pt-2">
                <Button type="submit" isLoading={passwordLoading}>Update Password</Button>
              </div>
            </form>
          </Card>

          <Card>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Two-Factor Authentication</h3>
            <p className="text-gray-500 dark:text-gray-400 mb-4">Add an extra layer of security to your account</p>
            <Button variant="outline">Enable 2FA</Button>
          </Card>
        </div>
      )}

      {/* API Keys Tab */}
      {activeTab === 'api-keys' && (
        <Card>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Binance API Keys</h3>
          <p className="text-sm text-gray-500 dark:text-gray-400 mb-4">
            Your API keys are encrypted at rest using AES-256-GCM. We never store your keys in plaintext.
          </p>
          <div className="space-y-4 max-w-lg">
            <Input label="API Key" type="password" placeholder="Enter your Binance API key" />
            <Input label="Secret Key" type="password" placeholder="Enter your Binance Secret key" />
            <div className="p-3 bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg">
              <p className="text-sm text-yellow-700 dark:text-yellow-400">
                <strong>Security Note:</strong> Only provide API keys with trading permissions. Never enable withdrawal permissions for copy trading.
              </p>
            </div>
            <div className="pt-2">
              <Button>Save API Keys</Button>
            </div>
          </div>
        </Card>
      )}

      {/* Notifications Tab */}
      {activeTab === 'notifications' && (
        <Card>
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">Notification Preferences</h3>
          <div className="space-y-4 max-w-lg">
            {[
              { key: 'email_on_trade' as const, label: 'Email on Trade Execution', desc: 'Receive email when a trade is executed' },
              { key: 'email_on_error' as const, label: 'Email on Errors', desc: 'Receive email when an order fails' },
              { key: 'push_notifications' as const, label: 'Push Notifications', desc: 'Browser push notifications for live updates' },
              { key: 'trade_alerts' as const, label: 'Trade Alerts', desc: 'In-app alerts for new trade signals' },
            ].map((item) => (
              <div key={item.key} className="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg">
                <div>
                  <p className="text-sm font-medium text-gray-900 dark:text-white">{item.label}</p>
                  <p className="text-xs text-gray-500 dark:text-gray-400">{item.desc}</p>
                </div>
                <button
                  onClick={() => setNotifications({ ...notifications, [item.key]: !notifications[item.key] })}
                  className={`relative w-11 h-6 rounded-full transition-colors ${
                    notifications[item.key] ? 'bg-blue-600' : 'bg-gray-300 dark:bg-gray-600'
                  }`}
                >
                  <span className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform ${
                    notifications[item.key] ? 'translate-x-5' : ''
                  }`} />
                </button>
              </div>
            ))}
            <div className="pt-2">
              <Button>Save Preferences</Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}
