import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { DemoAccount, Payment, Role } from './types';
import { demoAccounts, payments as seedPayments } from './data';
import { generatePaymentReference, TODAY } from './lib';

interface AppState {
  account: DemoAccount | null;
  payments: Payment[];
  theme: 'light' | 'dark';
  login: (account: DemoAccount) => void;
  logout: () => void;
  toggleTheme: () => void;
  recordPayment: (input: {
    tenantId: string;
    amount: number;
    method: Payment['method'];
    recordedBy: string;
  }) => void;
}

const AppContext = createContext<AppState | null>(null);

const STORAGE_KEY = 'rentledger-state-v1';

interface PersistedState {
  accountId: string | null;
  theme: 'light' | 'dark';
}

function loadPersisted(): PersistedState {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as PersistedState;
      return { accountId: parsed.accountId ?? null, theme: parsed.theme ?? 'light' };
    }
  } catch {
    // ignore
  }
  return { accountId: null, theme: 'light' };
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [persisted] = useState(loadPersisted);
  const [account, setAccount] = useState<DemoAccount | null>(() => {
    if (!persisted.accountId) return null;
    return demoAccounts.find((a) => a.id === persisted.accountId) ?? null;
  });
  const [payments, setPayments] = useState<Payment[]>(seedPayments);
  const [theme, setTheme] = useState<'light' | 'dark'>(persisted.theme);

  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'dark') root.classList.add('dark');
    else root.classList.remove('dark');
  }, [theme]);

  useEffect(() => {
    const state: PersistedState = {
      accountId: account?.id ?? null,
      theme,
    };
    try {
      localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch {
      // ignore
    }
  }, [account, theme]);

  const value = useMemo<AppState>(
    () => ({
      account,
      payments,
      theme,
      login: (acc) => setAccount(acc),
      logout: () => setAccount(null),
      toggleTheme: () => setTheme((t) => (t === 'light' ? 'dark' : 'light')),
      recordPayment: ({ tenantId, amount, method, recordedBy }) => {
        const newPayment: Payment = {
          id: `pay-${Date.now()}`,
          tenantId,
          amount,
          date: TODAY.toISOString().slice(0, 10),
          method,
          reference: generatePaymentReference(),
          recordedBy,
        };
        setPayments((prev) => [...prev, newPayment]);
      },
    }),
    [account, payments, theme],
  );

  return <AppContext.Provider value={value}>{children}</AppContext.Provider>;
}

export function useApp(): AppState {
  const ctx = useContext(AppContext);
  if (!ctx) throw new Error('useApp must be used within AppProvider');
  return ctx;
}

export function roleLabel(role: Role): string {
  switch (role) {
    case 'owner':
      return 'Owner';
    case 'manager':
      return 'Property Manager';
    case 'tenant':
      return 'Tenant';
  }
}
