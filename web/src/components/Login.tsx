import { useState } from 'react';
import { Building, Sun, Moon, ShieldCheck, UserCog, User, ChevronRight } from 'lucide-react';
import { useApp, roleLabel } from '../context';
import { demoAccounts } from '../data';
import type { Role } from '../types';

const ROLE_CONFIG: { role: Role; label: string; description: string; icon: typeof User }[] = [
  {
    role: 'owner',
    label: 'Property Owner',
    description: 'Full access to all properties, tenants, and payments',
    icon: ShieldCheck,
  },
  {
    role: 'manager',
    label: 'Property Manager',
    description: 'Manage assigned properties, tenants, and record payments',
    icon: UserCog,
  },
  {
    role: 'tenant',
    label: 'Tenant',
    description: 'View your lease, payment history, and make payments',
    icon: User,
  },
];

export function Login() {
  const { login, theme, toggleTheme } = useApp();
  const [selectedRole, setSelectedRole] = useState<Role | null>(null);

  const accounts = selectedRole ? demoAccounts.filter((a) => a.role === selectedRole) : [];

  return (
    <div className="min-h-screen flex flex-col bg-gray-50 dark:bg-gray-950">
      {/* Theme toggle */}
      <div className="absolute top-4 right-4">
        <button onClick={toggleTheme} className="btn-ghost p-2" aria-label="Toggle theme">
          {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
        </button>
      </div>

      <div className="flex-1 flex items-center justify-center p-4">
        <div className="w-full max-w-4xl">
          {/* Brand */}
          <div className="flex flex-col items-center mb-10">
            <div className="w-12 h-12 bg-brand-600 flex items-center justify-center mb-4">
              <Building size={24} className="text-white" />
            </div>
            <h1 className="text-2xl font-bold tracking-tight">RentLedger</h1>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              Property & Rent Management Dashboard
            </p>
          </div>

          {!selectedRole ? (
            <>
              <p className="text-center text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-6">
                Select your role to continue
              </p>
              <div className="grid sm:grid-cols-3 gap-4">
                {ROLE_CONFIG.map((r) => (
                  <button
                    key={r.role}
                    onClick={() => setSelectedRole(r.role)}
                    className="surface p-6 text-left hover:border-brand-400 dark:hover:border-brand-700 transition-colors group"
                  >
                    <div className="w-10 h-10 bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400 flex items-center justify-center mb-4 group-hover:bg-brand-100 dark:group-hover:bg-brand-900 transition-colors">
                      <r.icon size={20} />
                    </div>
                    <p className="font-semibold mb-1">{r.label}</p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 leading-relaxed">
                      {r.description}
                    </p>
                  </button>
                ))}
              </div>
            </>
          ) : (
            <>
              <div className="flex items-center gap-2 mb-6">
                <button
                  onClick={() => setSelectedRole(null)}
                  className="text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
                >
                  Roles
                </button>
                <ChevronRight size={14} className="text-gray-400" />
                <span className="text-sm font-semibold">{roleLabel(selectedRole)}</span>
              </div>

              <p className="text-center text-sm font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide mb-6">
                Choose a demo account
              </p>
              <div className="grid sm:grid-cols-2 gap-4 max-w-2xl mx-auto">
                {accounts.map((acc) => (
                  <button
                    key={acc.id}
                    onClick={() => login(acc)}
                    className="surface p-5 flex items-center gap-4 hover:border-brand-400 dark:hover:border-brand-700 transition-colors text-left"
                  >
                    <div className="w-11 h-11 bg-brand-100 text-brand-700 dark:bg-brand-950 dark:text-brand-300 flex items-center justify-center font-semibold">
                      {acc.name
                        .split(' ')
                        .map((n) => n[0])
                        .slice(0, 2)
                        .join('')}
                    </div>
                    <div className="min-w-0">
                      <p className="font-semibold text-sm">{acc.name}</p>
                      <p className="text-xs text-gray-500 dark:text-gray-400 truncate">
                        {roleLabel(acc.role)}
                        {acc.tenantId && ' · Demo Tenant'}
                        {acc.managerId && ' · Demo Manager'}
                      </p>
                    </div>
                    <ChevronRight size={18} className="text-gray-300 dark:text-gray-600 ml-auto" />
                  </button>
                ))}
              </div>
            </>
          )}

          <p className="text-center text-xs text-gray-400 dark:text-gray-600 mt-10">
            Demo environment — all data is sample data. No real authentication is performed.
          </p>
        </div>
      </div>
    </div>
  );
}
