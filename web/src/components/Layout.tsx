import { useState, type ReactNode } from 'react';
import {
  LayoutDashboard,
  Building2,
  Users,
  Receipt,
  CalendarClock,
  LogOut,
  Sun,
  Moon,
  Menu,
  X,
  Building,
  UserPlus,
} from 'lucide-react';
import { useApp, roleLabel } from '../context';
import type { Role } from '../types';

export type Page = 'overview' | 'properties' | 'tenants' | 'payments' | 'due-dates' | 'tenant-home' | 'create-account';

interface NavItem {
  id: Page;
  label: string;
  icon: typeof LayoutDashboard;
  roles: Role[];
}

const NAV_ITEMS: NavItem[] = [
  { id: 'overview', label: 'Overview', icon: LayoutDashboard, roles: ['owner', 'manager'] },
  { id: 'properties', label: 'Properties', icon: Building2, roles: ['owner', 'manager'] },
  { id: 'tenants', label: 'Tenants', icon: Users, roles: ['owner', 'manager'] },
  { id: 'payments', label: 'Payments', icon: Receipt, roles: ['owner', 'manager'] },
  { id: 'due-dates', label: 'Due Dates', icon: CalendarClock, roles: ['owner', 'manager'] },
  { id: 'create-account', label: 'Create Account', icon: UserPlus, roles: ['owner', 'manager'] },
  { id: 'tenant-home', label: 'My Account', icon: LayoutDashboard, roles: ['tenant'] },
];

export function Layout({
  current,
  onNavigate,
  children,
}: {
  current: Page;
  onNavigate: (p: Page) => void;
  children: ReactNode;
}) {
  const { account, logout, theme, toggleTheme } = useApp();
  const [mobileOpen, setMobileOpen] = useState(false);

  if (!account) return null;

  const items = NAV_ITEMS.filter((i) => i.roles.includes(account.role));

  const sidebar = (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-2.5 px-5 py-5 border-b border-gray-200 dark:border-gray-800">
        <div className="w-8 h-8 bg-brand-600 flex items-center justify-center">
          <Building size={18} className="text-white" />
        </div>
        <div>
          <p className="font-bold text-sm leading-tight">RentLedger</p>
          <p className="text-[10px] text-gray-400 uppercase tracking-wider">Property Mgmt</p>
        </div>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-0.5 overflow-y-auto">
        {items.map((item) => {
          const active = current === item.id;
          return (
            <button
              key={item.id}
              onClick={() => {
                onNavigate(item.id);
                setMobileOpen(false);
              }}
              className={`sidebar-link w-full text-left ${active ? 'sidebar-link-active' : ''}`}
            >
              <item.icon size={18} />
              {item.label}
            </button>
          );
        })}
      </nav>

      <div className="px-3 py-4 border-t border-gray-200 dark:border-gray-800">
        <div className="flex items-center gap-3 px-3 py-2">
          <div className="w-9 h-9 bg-brand-100 text-brand-700 dark:bg-brand-950 dark:text-brand-300 flex items-center justify-center font-semibold text-sm">
            {account.name
              .split(' ')
              .map((n) => n[0])
              .slice(0, 2)
              .join('')}
          </div>
          <div className="min-w-0 flex-1">
            <p className="text-sm font-medium truncate">{account.name}</p>
            <p className="text-xs text-gray-400 truncate">{roleLabel(account.role)}</p>
          </div>
        </div>
        <button
          onClick={logout}
          className="sidebar-link w-full text-left mt-1 text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-950"
        >
          <LogOut size={18} />
          Sign Out
        </button>
      </div>
    </div>
  );

  return (
    <div className="min-h-screen flex bg-gray-50 dark:bg-gray-950">
      {/* Desktop sidebar */}
      <aside className="hidden md:flex w-60 shrink-0 surface border-r border-gray-200 dark:border-gray-800 sticky top-0 h-screen">
        {sidebar}
      </aside>

      {/* Mobile sidebar */}
      {mobileOpen && (
        <div className="md:hidden fixed inset-0 z-50 flex">
          <div className="absolute inset-0 bg-black/40" onClick={() => setMobileOpen(false)} />
          <aside className="relative w-60 surface border-r border-gray-200 dark:border-gray-800 h-full">
            {sidebar}
          </aside>
        </div>
      )}

      <div className="flex-1 min-w-0 flex flex-col">
        {/* Top bar */}
        <header className="sticky top-0 z-30 surface border-b border-gray-200 dark:border-gray-800 px-4 md:px-6 py-3 flex items-center justify-between">
          <div className="flex items-center gap-3">
            <button
              className="md:hidden btn-ghost p-1.5"
              onClick={() => setMobileOpen(true)}
              aria-label="Open menu"
            >
              <Menu size={20} />
            </button>
            <div className="hidden sm:block">
              <p className="text-xs text-gray-400">
                {new Date().toLocaleDateString('en-US', { weekday: 'long', month: 'long', day: 'numeric' })}
              </p>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={toggleTheme}
              className="btn-ghost p-2"
              aria-label="Toggle theme"
              title="Toggle theme"
            >
              {theme === 'light' ? <Moon size={18} /> : <Sun size={18} />}
            </button>
            <div className="w-px h-6 bg-gray-200 dark:bg-gray-800" />
            <div className="flex items-center gap-2.5">
              <div className="w-8 h-8 bg-brand-100 text-brand-700 dark:bg-brand-950 dark:text-brand-300 flex items-center justify-center font-semibold text-xs">
                {account.name
                  .split(' ')
                  .map((n) => n[0])
                  .slice(0, 2)
                  .join('')}
              </div>
              <div className="hidden sm:block">
                <p className="text-sm font-medium leading-tight">{account.name}</p>
                <p className="text-xs text-gray-400 leading-tight">{roleLabel(account.role)}</p>
              </div>
            </div>
          </div>
        </header>

        <main className="flex-1 p-4 md:p-6 max-w-7xl w-full mx-auto">{children}</main>
      </div>

      {mobileOpen && (
        <button
          className="md:hidden fixed top-3 right-3 z-50 btn-ghost p-2"
          onClick={() => setMobileOpen(false)}
        >
          <X size={20} />
        </button>
      )}
    </div>
  );
}
