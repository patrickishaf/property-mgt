import type { LeaseStatus, Payment, Property, Tenant, Unit, PaymentStatus } from './types';
import type { ReceiptData } from './components/ReceiptModal';
import { payments as seedPayments, properties, tenants, units } from './data';

export const CYCLE_YEAR = 2025;
export const CYCLE_MONTH = 7; // August (0-indexed)
export const TODAY = new Date(2025, 7, 12); // Aug 12, 2025 — fixed reference for consistent statuses

const MONTH_NAMES = [
  'January', 'February', 'March', 'April', 'May', 'June',
  'July', 'August', 'September', 'October', 'November', 'December',
];

export function formatCurrency(amount: number): string {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'GHS',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

export function formatDate(iso: string): string {
  const d = new Date(iso);
  return `${d.getDate()} ${MONTH_NAMES[d.getMonth()].slice(0, 3)} ${d.getFullYear()}`;
}

export function formatDateShort(iso: string): string {
  const d = new Date(iso);
  return `${d.getDate()} ${MONTH_NAMES[d.getMonth()].slice(0, 3)}`;
}

export function computeNextDueDate(dueDay: number, from: Date = TODAY): string {
  const d = new Date(from.getFullYear(), from.getMonth(), dueDay);
  if (d < from) {
    d.setMonth(d.getMonth() + 1);
  }
  return d.toISOString().slice(0, 10);
}

export function daysBetween(fromIso: string, toIso: string): number {
  const a = new Date(fromIso).getTime();
  const b = new Date(toIso).getTime();
  return Math.round((b - a) / (1000 * 60 * 60 * 24));
}

function getStatus(totalPaid: number, rent: number, daysUntilDue: number): PaymentStatus {
  if (totalPaid >= rent) return 'paid';
  if (daysUntilDue < 0) return 'overdue';
  return 'partial';
}

export function buildLeaseStatus(
  tenant: Tenant,
  allPayments: Payment[],
  allUnits: Unit[],
  allProperties: Property[],
  asOf: Date = TODAY,
): LeaseStatus {
  const unit = allUnits.find((u) => u.id === tenant.unitId)!;
  const property = allProperties.find((p) => p.id === tenant.propertyId)!;
  const cycleStart = new Date(asOf.getFullYear(), asOf.getMonth(), 1).toISOString().slice(0, 10);
  const tenantPayments = allPayments
    .filter((p) => p.tenantId === tenant.id && p.date >= cycleStart)
    .sort((a, b) => a.date.localeCompare(b.date));
  const totalPaid = tenantPayments.reduce((sum, p) => sum + p.amount, 0);
  const balance = Math.max(0, tenant.monthlyRent - totalPaid);
  const nextDueDate = computeNextDueDate(tenant.dueDayOfMonth, asOf);
  const daysUntilDue = daysBetween(asOf.toISOString().slice(0, 10), nextDueDate);
  const status = getStatus(totalPaid, tenant.monthlyRent, daysUntilDue);
  return {
    tenant,
    unit,
    property,
    monthlyRent: tenant.monthlyRent,
    totalPaidThisCycle: totalPaid,
    balance,
    status,
    nextDueDate,
    daysUntilDue,
  };
}

export function buildAllLeaseStatuses(allPayments: Payment[] = seedPayments): LeaseStatus[] {
  return tenants.map((t) => buildLeaseStatus(t, allPayments, units, properties));
}

export function getTenantPayments(tenantId: string, allPayments: Payment[]): Payment[] {
  return allPayments
    .filter((p) => p.tenantId === tenantId)
    .sort((a, b) => b.date.localeCompare(a.date));
}

export function getCyclePayments(tenantId: string, allPayments: Payment[]): Payment[] {
  const cycleStart = new Date(TODAY.getFullYear(), TODAY.getMonth(), 1).toISOString().slice(0, 10);
  return allPayments
    .filter((p) => p.tenantId === tenantId && p.date >= cycleStart)
    .sort((a, b) => a.date.localeCompare(b.date));
}

export function getProperty(propertyId: string): Property | undefined {
  return properties.find((p) => p.id === propertyId);
}

export function getUnit(unitId: string): Unit | undefined {
  return units.find((u) => u.id === unitId);
}

export function getTenant(tenantId: string): Tenant | undefined {
  return tenants.find((t) => t.id === tenantId);
}

export function getPropertyTenants(propertyId: string): Tenant[] {
  return tenants.filter((t) => t.propertyId === propertyId);
}

export function getVacantUnits(): Unit[] {
  return units.filter((u) => !tenants.some((t) => t.unitId === u.id));
}

export function generatePaymentReference(): string {
  const n = Math.floor(Math.random() * 90000 + 10000);
  return `PAY${n}`;
}

export function buildReceiptData(payment: Payment, allPayments: Payment[]): ReceiptData {
  const tenant = tenants.find((t) => t.id === payment.tenantId);
  const unit = units.find((u) => u.id === tenant?.unitId);
  const property = properties.find((p) => p.id === tenant?.propertyId);
  if (!tenant || !unit || !property) throw new Error('Missing data for receipt');

  const cycleStart = new Date(TODAY.getFullYear(), TODAY.getMonth(), 1).toISOString().slice(0, 10);
  const cyclePayments = allPayments
    .filter((p) => p.tenantId === tenant.id && p.date >= cycleStart)
    .sort((a, b) => a.date.localeCompare(b.date));
  const totalPaidThisCycle = cyclePayments.reduce((s, p) => s + p.amount, 0);
  const balance = Math.max(0, tenant.monthlyRent - totalPaidThisCycle);
  const paymentNumber = cyclePayments.findIndex((p) => p.id === payment.id) + 1;
  const paymentType = payment.amount >= tenant.monthlyRent ? 'full' : 'partial';

  return {
    payment,
    tenant,
    property,
    unit,
    monthlyRent: tenant.monthlyRent,
    totalPaidThisCycle,
    balance,
    paymentType,
    paymentNumber: paymentNumber || 1,
  };
}

