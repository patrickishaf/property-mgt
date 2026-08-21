import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react';
import type { DemoAccount, Payment, Role } from './types';
import { payments as seedPayments } from './data';
import { generatePaymentReference, TODAY } from './lib';
import { mockAuth } from './mockAuth';

interface AppState {
  account: DemoAccount | null;
  authReady: boolean;
  authLoading: boolean;
  payments: Payment[];
  theme: 'light' | 'dark';
  loginWithPassword: (identifier: string, password: string) => Promise<void>;
  sendOtp: (identifier: string) => Promise<string>;
  verifyOtp: (identifier: string, code: string) => Promise<void>;
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

const THEME_KEY = 'rentledger-theme';

function loadTheme(): 'light' | 'dark' {
  try {
    const raw = localStorage.getItem(THEME_KEY);
    if (raw === 'dark' || raw === 'light') return raw;
  } catch {
    // ignore
  }
  return 'light';
}

export function AppProvider({ children }: { children: ReactNode }) {
  const [account, setAccount] = useState<DemoAccount | null>(null);
  const [authReady, setAuthReady] = useState(false);
  const [authLoading, setAuthLoading] = useState(false);
  const [payments, setPayments] = useState<Payment[]>(seedPayments);
  const [theme, setTheme] = useState<'light' | 'dark'>(loadTheme);

  useEffect(() => {
    const root = document.documentElement;
    if (theme === 'dark') root.classList.add('dark');
    else root.classList.remove('dark');
    try {
      localStorage.setItem(THEME_KEY, theme);
    } catch {
      // ignore
    }
  }, [theme]);

  useEffect(() => {
    const sessionId = mockAuth.getSession();
    if (sessionId) {
      const acc = mockAuth.getAccountById(sessionId);
      if (acc) setAccount(acc);
      else mockAuth.clearSession();
    }
    setAuthReady(true);
  }, []);

  const value = useMemo<AppState>(
    () => ({
      account,
      authReady,
      authLoading,
      payments,
      theme,
      loginWithPassword: async (identifier, password) => {
        setAuthLoading(true);
        try {
          const acc = await mockAuth.loginWithPassword(identifier, password);
          setAccount(acc);
        } finally {
          setAuthLoading(false);
        }
      },
      sendOtp: async (identifier) => {
        setAuthLoading(true);
        try {
          const { code } = await mockAuth.sendOtp(identifier);
          return code;
        } finally {
          setAuthLoading(false);
        }
      },
      verifyOtp: async (identifier, code) => {
        setAuthLoading(true);
        try {
          const acc = await mockAuth.verifyOtp(identifier, code);
          setAccount(acc);
        } finally {
          setAuthLoading(false);
        }
      },
      logout: () => {
        mockAuth.clearSession();
        setAccount(null);
      },
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
    [account, authReady, authLoading, payments, theme],
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
