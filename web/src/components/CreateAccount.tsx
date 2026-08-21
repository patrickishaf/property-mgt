import { useState } from 'react';
import {
  UserPlus,
  User,
  UserCog,
  Check,
  Loader2,
  AlertCircle,
  Mail,
  Phone,
  Lock,
  Building2,
  Home,
} from 'lucide-react';
import { useApp } from '../context';
import { SectionHeader, Card } from './ui';
import { properties, units, tenants } from '../data';
import { mockAuth } from '../mockAuth';
import type { Role } from '../types';

type FormRole = 'tenant' | 'manager';

export function CreateAccount() {
  const { account } = useApp();
  const [role, setRole] = useState<FormRole>('tenant');
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [phone, setPhone] = useState('');
  const [password, setPassword] = useState('');
  const [propertyId, setPropertyId] = useState('');
  const [unitId, setUnitId] = useState('');
  const [monthlyRent, setMonthlyRent] = useState('');
  const [dueDay, setDueDay] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [success, setSuccess] = useState<string | null>(null);

  const canCreateManager = account?.role === 'owner';

  const availableUnits = units.filter(
    (u) => u.propertyId === propertyId && !tenants.some((t) => t.unitId === u.id),
  );

  const handleRoleChange = (r: FormRole) => {
    setRole(r);
    setError('');
    setSuccess(null);
    setPropertyId('');
    setUnitId('');
    setMonthlyRent('');
    setDueDay('');
  };

  const handleSubmit = async () => {
    setError('');
    setSuccess(null);

    if (!name.trim()) return setError('Enter the full name.');
    if (!email.trim()) return setError('Enter an email address.');
    if (!phone.trim()) return setError('Enter a phone number.');
    if (!password || password.length < 6)
      return setError('Password must be at least 6 characters.');

    if (role === 'tenant') {
      if (!propertyId) return setError('Select a property.');
      if (!unitId) return setError('Select a unit.');
      if (!monthlyRent || parseFloat(monthlyRent) <= 0)
        return setError('Enter a valid monthly rent.');
      if (!dueDay || parseInt(dueDay) < 1 || parseInt(dueDay) > 31)
        return setError('Enter a due day between 1 and 31.');
    }

    setLoading(true);
    try {
      const input = {
        name: name.trim(),
        email: email.trim(),
        phone: phone.trim(),
        password,
        role: role as Role,
        ...(role === 'tenant' ? { tenantId: `t-${Date.now()}` } : { managerId: `mgr-${Date.now()}` }),
      };
      const newAccount = await mockAuth.createAccount(input);
      setSuccess(
        `Account created for ${newAccount.name}. They can now sign in using ${newAccount.email} or ${newAccount.phone}.`,
      );
      setName('');
      setEmail('');
      setPhone('');
      setPassword('');
      setPropertyId('');
      setUnitId('');
      setMonthlyRent('');
      setDueDay('');
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to create account.');
    } finally {
      setLoading(false);
    }
  };

  if (success) {
    return (
      <div>
        <SectionHeader title="Create Account" subtitle="Create a new tenant or manager account" />
        <Card className="p-8">
          <div className="flex flex-col items-center text-center">
            <div className="w-14 h-14 bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400 flex items-center justify-center mb-4">
              <Check size={28} />
            </div>
            <p className="text-lg font-semibold mb-2">Account Created Successfully</p>
            <p className="text-sm text-gray-500 dark:text-gray-400 max-w-sm mb-6">{success}</p>
            <button
              className="btn btn-primary"
              onClick={() => setSuccess(null)}
            >
              <UserPlus size={16} /> Create Another Account
            </button>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div>
      <SectionHeader
        title="Create Account"
        subtitle="Create a new tenant or manager account"
      />

      <div className="max-w-2xl">
        {/* Role selector */}
        <div className="grid grid-cols-2 gap-3 mb-6">
          <button
            onClick={() => handleRoleChange('tenant')}
            className={`surface p-4 flex items-center gap-3 transition-colors text-left ${
              role === 'tenant'
                ? 'border-brand-600 dark:border-brand-500'
                : 'hover:border-gray-300 dark:hover:border-gray-700'
            }`}
          >
            <div
              className={`w-10 h-10 flex items-center justify-center ${
                role === 'tenant'
                  ? 'bg-brand-600 text-white'
                  : 'bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400'
              }`}
            >
              <User size={18} />
            </div>
            <div>
              <p className="font-semibold text-sm">Tenant</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">Access lease & payments</p>
            </div>
          </button>

          <button
            onClick={() => handleRoleChange('manager')}
            disabled={!canCreateManager}
            className={`surface p-4 flex items-center gap-3 transition-colors text-left disabled:opacity-50 disabled:cursor-not-allowed ${
              role === 'manager'
                ? 'border-brand-600 dark:border-brand-500'
                : 'hover:border-gray-300 dark:hover:border-gray-700'
            }`}
          >
            <div
              className={`w-10 h-10 flex items-center justify-center ${
                role === 'manager'
                  ? 'bg-brand-600 text-white'
                  : 'bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400'
              }`}
            >
              <UserCog size={18} />
            </div>
            <div>
              <p className="font-semibold text-sm">Manager</p>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                {canCreateManager ? 'Manage properties & tenants' : 'Owner only'}
              </p>
            </div>
          </button>
        </div>

        {!canCreateManager && role === 'manager' && (
          <div className="flex items-start gap-2 px-3 py-2.5 text-sm text-amber-700 bg-amber-50 border border-amber-200 dark:text-amber-300 dark:bg-amber-950 dark:border-amber-800 mb-4">
            <AlertCircle size={16} className="shrink-0 mt-0.5" />
            <span>Only the property owner can create new manager accounts.</span>
          </div>
        )}

        <Card className="p-5 sm:p-6">
          {/* Error */}
          {error && (
            <div className="flex items-start gap-2 px-3 py-2.5 text-sm text-red-700 bg-red-50 border border-red-200 dark:text-red-300 dark:bg-red-950 dark:border-red-800 mb-4">
              <AlertCircle size={16} className="shrink-0 mt-0.5" />
              <span>{error}</span>
            </div>
          )}

          {/* Name */}
          <div className="mb-4">
            <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
              Full Name
            </label>
            <input
              type="text"
              className="input-base w-full px-3 py-2.5 text-sm"
              placeholder="e.g. Kofi Mensah"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </div>

          {/* Email + Phone */}
          <div className="grid sm:grid-cols-2 gap-4 mb-4">
            <div>
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Email
              </label>
              <div className="relative">
                <Mail size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="email"
                  className="input-base w-full pl-9 pr-3 py-2.5 text-sm"
                  placeholder="name@example.com"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>
            <div>
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Phone
              </label>
              <div className="relative">
                <Phone size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
                <input
                  type="tel"
                  className="input-base w-full pl-9 pr-3 py-2.5 text-sm"
                  placeholder="+233 24 000 0000"
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                />
              </div>
            </div>
          </div>

          {/* Password */}
          <div className="mb-4">
            <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
              Temporary Password
            </label>
            <div className="relative">
              <Lock size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
              <input
                type="password"
                className="input-base w-full pl-9 pr-3 py-2.5 text-sm"
                placeholder="At least 6 characters"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </div>
            <p className="text-xs text-gray-400 mt-1.5">
              The user will be asked to change this on first login.
            </p>
          </div>

          {/* Tenant-specific fields */}
          {role === 'tenant' && (
            <div className="pt-4 mt-4 border-t border-gray-100 dark:border-gray-800">
              <div className="flex items-center gap-2 mb-4">
                <Building2 size={16} className="text-gray-400" />
                <p className="text-sm font-semibold">Lease Assignment</p>
              </div>

              <div className="grid sm:grid-cols-2 gap-4 mb-4">
                <div>
                  <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                    Property
                  </label>
                  <select
                    className="input-base w-full px-3 py-2.5 text-sm"
                    value={propertyId}
                    onChange={(e) => {
                      setPropertyId(e.target.value);
                      setUnitId('');
                    }}
                  >
                    <option value="">Select a property…</option>
                    {properties.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                    Unit
                  </label>
                  <select
                    className="input-base w-full px-3 py-2.5 text-sm"
                    value={unitId}
                    onChange={(e) => setUnitId(e.target.value)}
                    disabled={!propertyId}
                  >
                    <option value="">Select a unit…</option>
                    {availableUnits.map((u) => (
                      <option key={u.id} value={u.id}>
                        {u.label} (Floor {u.floor})
                      </option>
                    ))}
                  </select>
                  {propertyId && availableUnits.length === 0 && (
                    <p className="text-xs text-amber-500 mt-1">No vacant units in this property.</p>
                  )}
                </div>
              </div>

              <div className="grid sm:grid-cols-2 gap-4 mb-4">
                <div>
                  <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                    Monthly Rent (GHS)
                  </label>
                  <input
                    type="number"
                    className="input-base w-full px-3 py-2.5 text-sm"
                    placeholder="0"
                    value={monthlyRent}
                    onChange={(e) => setMonthlyRent(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                    Due Day of Month
                  </label>
                  <input
                    type="number"
                    min={1}
                    max={31}
                    className="input-base w-full px-3 py-2.5 text-sm"
                    placeholder="e.g. 5"
                    value={dueDay}
                    onChange={(e) => setDueDay(e.target.value)}
                  />
                </div>
              </div>
            </div>
          )}

          {/* Manager-specific note */}
          {role === 'manager' && (
            <div className="pt-4 mt-4 border-t border-gray-100 dark:border-gray-800">
              <div className="flex items-center gap-2 mb-2">
                <Home size={16} className="text-gray-400" />
                <p className="text-sm font-semibold">Property Assignment</p>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400">
                After creating the account, you can assign properties to this manager from the
                Properties page.
              </p>
            </div>
          )}

          {/* Submit */}
          <div className="flex items-center justify-end gap-3 pt-5 mt-5 border-t border-gray-100 dark:border-gray-800">
            <button
              onClick={handleSubmit}
              disabled={loading || (role === 'manager' && !canCreateManager)}
              className="btn btn-primary"
            >
              {loading ? <Loader2 size={16} className="animate-spin" /> : <UserPlus size={16} />}
              {loading ? 'Creating…' : 'Create Account'}
            </button>
          </div>
        </Card>
      </div>
    </div>
  );
}
