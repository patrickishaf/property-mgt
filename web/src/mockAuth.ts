import type { DemoAccount, Role } from './types';
import { demoAccounts } from './data';

const ACCOUNTS_KEY = 'rentledger-accounts-v1';
const SESSION_KEY = 'rentledger-session-v1';

interface StoredAccount extends DemoAccount {}

interface OtpRecord {
  identifier: string;
  code: string;
  expiresAt: number;
}

function delay(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function loadAccounts(): StoredAccount[] {
  try {
    const raw = localStorage.getItem(ACCOUNTS_KEY);
    if (raw) {
      const parsed = JSON.parse(raw) as StoredAccount[];
      if (Array.isArray(parsed) && parsed.length > 0) return parsed;
    }
  } catch {
    // ignore
  }
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(demoAccounts));
  return [...demoAccounts];
}

function saveAccounts(accounts: StoredAccount[]): void {
  localStorage.setItem(ACCOUNTS_KEY, JSON.stringify(accounts));
}

let activeOtp: OtpRecord | null = null;

function generateOtp(): string {
  return String(Math.floor(100000 + Math.random() * 900000));
}

export interface CreateAccountInput {
  name: string;
  email: string;
  phone: string;
  role: Role;
  password: string;
  tenantId?: string;
  managerId?: string;
}

export const mockAuth = {
  async findAccount(identifier: string): Promise<StoredAccount | null> {
    await delay(400);
    const accounts = loadAccounts();
    const id = identifier.trim().toLowerCase();
    return (
      accounts.find(
        (a) => a.email.toLowerCase() === id || a.phone.replace(/\s/g, '') === identifier.trim().replace(/\s/g, ''),
      ) ?? null
    );
  },

  async loginWithPassword(identifier: string, password: string): Promise<StoredAccount> {
    await delay(600);
    const account = await this.findAccount(identifier);
    if (!account) throw new Error('No account found with that email or phone number.');
    if (account.password !== password) throw new Error('Incorrect password. Please try again.');
    localStorage.setItem(SESSION_KEY, JSON.stringify({ id: account.id }));
    return account;
  },

  async sendOtp(identifier: string): Promise<{ code: string }> {
    await delay(500);
    const account = await this.findAccount(identifier);
    if (!account) throw new Error('No account found with that email or phone number.');
    const code = generateOtp();
    activeOtp = { identifier: identifier.trim().toLowerCase(), code, expiresAt: Date.now() + 5 * 60 * 1000 };
    return { code };
  },

  async verifyOtp(identifier: string, code: string): Promise<StoredAccount> {
    await delay(500);
    if (!activeOtp) throw new Error('No code was sent. Please request a new one.');
    if (Date.now() > activeOtp.expiresAt) {
      activeOtp = null;
      throw new Error('This code has expired. Please request a new one.');
    }
    if (activeOtp.identifier !== identifier.trim().toLowerCase()) {
      throw new Error('This code was sent to a different account.');
    }
    if (activeOtp.code !== code) throw new Error('Incorrect code. Please try again.');
    activeOtp = null;
    const account = await this.findAccount(identifier);
    if (!account) throw new Error('Account not found.');
    localStorage.setItem(SESSION_KEY, JSON.stringify({ id: account.id }));
    return account;
  },

  async createAccount(input: CreateAccountInput): Promise<StoredAccount> {
    await delay(800);
    const accounts = loadAccounts();
    const exists = accounts.find(
      (a) =>
        a.email.toLowerCase() === input.email.toLowerCase() ||
        a.phone.replace(/\s/g, '') === input.phone.replace(/\s/g, ''),
    );
    if (exists) throw new Error('An account with this email or phone already exists.');
    const id = `acc-${Date.now()}`;
    const newAccount: StoredAccount = {
      id,
      role: input.role,
      name: input.name,
      email: input.email,
      phone: input.phone,
      password: input.password,
      tenantId: input.tenantId,
      managerId: input.managerId,
    };
    accounts.push(newAccount);
    saveAccounts(accounts);
    return newAccount;
  },

  getSession(): string | null {
    try {
      const raw = localStorage.getItem(SESSION_KEY);
      if (!raw) return null;
      const parsed = JSON.parse(raw) as { id: string };
      return parsed.id ?? null;
    } catch {
      return null;
    }
  },

  getAccountById(id: string): StoredAccount | null {
    const accounts = loadAccounts();
    return accounts.find((a) => a.id === id) ?? null;
  },

  clearSession(): void {
    localStorage.removeItem(SESSION_KEY);
  },
};
