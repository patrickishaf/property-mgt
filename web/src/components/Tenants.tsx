import { useMemo, useState } from 'react';
import { Search, ArrowLeft, Plus, Mail, Phone, Receipt } from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, StatusBadge, Avatar, EmptyState, ProgressBar } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import { ReceiptModal, type ReceiptData } from './ReceiptModal';
import { Printer } from 'lucide-react';
import {
  buildAllLeaseStatuses,
  buildReceiptData,
  formatCurrency,
  formatDate,
  formatDateShort,
  getCyclePayments,
} from '../lib';
import { properties, tenants as allTenants } from '../data';
import type { PaymentStatus } from '../types';

export function TenantsList({ onSelectTenant }: { onSelectTenant: (id: string) => void }) {
  const { payments } = useApp();
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<PaymentStatus | 'all'>('all');
  const [propertyFilter, setPropertyFilter] = useState<string>('all');

  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);

  const filtered = statuses.filter((s) => {
    const matchesSearch =
      !search ||
      s.tenant.name.toLowerCase().includes(search.toLowerCase()) ||
      s.tenant.email.toLowerCase().includes(search.toLowerCase()) ||
      s.unit.label.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === 'all' || s.status === statusFilter;
    const matchesProperty = propertyFilter === 'all' || s.property.id === propertyFilter;
    return matchesSearch && matchesStatus && matchesProperty;
  });

  return (
    <div>
      <SectionHeader title="Tenants" subtitle={`${allTenants.length} tenants across all properties`} />

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-5">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search by name, email, or unit…"
            className="input-base w-full pl-9 pr-3 py-2 text-sm"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <select
          className="input-base px-3 py-2 text-sm"
          value={propertyFilter}
          onChange={(e) => setPropertyFilter(e.target.value)}
        >
          <option value="all">All Properties</option>
          {properties.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name}
            </option>
          ))}
        </select>
        <select
          className="input-base px-3 py-2 text-sm"
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as PaymentStatus | 'all')}
        >
          <option value="all">All Statuses</option>
          <option value="paid">Paid</option>
          <option value="partial">Partial</option>
          <option value="overdue">Overdue</option>
        </select>
      </div>

      {/* Table */}
      <Card>
        {filtered.length === 0 ? (
          <EmptyState
            icon={<Search size={32} />}
            title="No tenants found"
            message="Try adjusting your search or filters."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-gray-400 uppercase tracking-wide border-b border-gray-100 dark:border-gray-800">
                  <th className="text-left font-medium px-5 py-3">Tenant</th>
                  <th className="text-left font-medium px-5 py-3">Property</th>
                  <th className="text-left font-medium px-5 py-3">Unit</th>
                  <th className="text-left font-medium px-5 py-3">Rent</th>
                  <th className="text-left font-medium px-5 py-3">Balance</th>
                  <th className="text-left font-medium px-5 py-3">Next Due</th>
                  <th className="text-left font-medium px-5 py-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                {filtered.map((s) => (
                  <tr
                    key={s.tenant.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer"
                    onClick={() => onSelectTenant(s.tenant.id)}
                  >
                    <td className="px-5 py-3">
                      <div className="flex items-center gap-3">
                        <Avatar name={s.tenant.name} size={32} />
                        <div className="min-w-0">
                          <p className="font-medium truncate">{s.tenant.name}</p>
                          <p className="text-xs text-gray-400 truncate">{s.tenant.email}</p>
                        </div>
                      </div>
                    </td>
                    <td className="px-5 py-3 text-gray-500">{s.property.name}</td>
                    <td className="px-5 py-3 text-gray-500">{s.unit.label}</td>
                    <td className="px-5 py-3">{formatCurrency(s.monthlyRent)}</td>
                    <td className="px-5 py-3">
                      {s.balance > 0 ? (
                        <span className="text-amber-600 dark:text-amber-400 font-medium">
                          {formatCurrency(s.balance)}
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-5 py-3 text-gray-500">{formatDate(s.nextDueDate)}</td>
                    <td className="px-5 py-3"><StatusBadge status={s.status} /></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </Card>
    </div>
  );
}

export function TenantDetail({
  tenantId,
  onBack,
}: {
  tenantId: string;
  onBack: () => void;
}) {
  const { account, payments } = useApp();
  const [showRecord, setShowRecord] = useState(false);
  const [receiptData, setReceiptData] = useState<ReceiptData | null>(null);

  const tenant = allTenants.find((t) => t.id === tenantId);
  const status = useMemo(
    () => buildAllLeaseStatuses(payments).find((s) => s.tenant.id === tenantId),
    [payments, tenantId],
  );
  const cyclePayments = useMemo(
    () => (status ? getCyclePayments(tenantId, payments) : []),
    [status, tenantId, payments],
  );

  if (!tenant || !status) return null;

  const pct = (status.totalPaidThisCycle / status.monthlyRent) * 100;

  return (
    <div>
      <button
        onClick={onBack}
        className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 mb-4"
      >
        <ArrowLeft size={16} /> Back to Tenants
      </button>

      {/* Tenant header */}
      <div className="surface p-5 mb-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div className="flex items-center gap-4">
            <Avatar name={tenant.name} size={56} />
            <div>
              <h1 className="text-xl font-bold">{tenant.name}</h1>
              <div className="flex flex-wrap items-center gap-x-4 gap-y-1 mt-1 text-xs text-gray-500 dark:text-gray-400">
                <span className="flex items-center gap-1"><Mail size={12} /> {tenant.email}</span>
                <span className="flex items-center gap-1"><Phone size={12} /> {tenant.phone}</span>
              </div>
            </div>
          </div>
          <button className="btn btn-primary" onClick={() => setShowRecord(true)}>
            <Plus size={16} /> Record Payment
          </button>
        </div>
      </div>

      {/* Lease summary */}
      <div className="grid lg:grid-cols-3 gap-6 mb-6">
        <Card className="p-5 lg:col-span-2">
          <p className="text-sm font-semibold mb-4">Lease & Payment Summary</p>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-5">
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Property</p>
              <p className="text-sm font-medium mt-1">{status.property.name}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Unit</p>
              <p className="text-sm font-medium mt-1">{status.unit.label}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Lease Start</p>
              <p className="text-sm font-medium mt-1">{formatDate(tenant.leaseStart)}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Due Day</p>
              <p className="text-sm font-medium mt-1">{tenant.dueDayOfMonth}th of month</p>
            </div>
          </div>

          <div className="pt-4 border-t border-gray-100 dark:border-gray-800">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {formatCurrency(status.totalPaidThisCycle)} of {formatCurrency(status.monthlyRent)} collected
              </span>
              <span className="text-sm font-bold">{pct.toFixed(0)}%</span>
            </div>
            <ProgressBar value={status.totalPaidThisCycle} max={status.monthlyRent} />
          </div>
        </Card>

        <div className="grid grid-cols-2 lg:grid-cols-1 gap-4">
          <Card className="p-4">
            <p className="text-xs text-gray-400 uppercase tracking-wide">Monthly Rent</p>
            <p className="text-xl font-bold mt-1">{formatCurrency(status.monthlyRent)}</p>
          </Card>
          <Card className="p-4">
            <p className="text-xs text-gray-400 uppercase tracking-wide">Outstanding Balance</p>
            <p className="text-xl font-bold mt-1 text-amber-600 dark:text-amber-400">
              {formatCurrency(status.balance)}
            </p>
          </Card>
          <Card className="p-4">
            <p className="text-xs text-gray-400 uppercase tracking-wide">Next Due Date</p>
            <p className="text-xl font-bold mt-1">{formatDate(status.nextDueDate)}</p>
          </Card>
          <Card className="p-4">
            <p className="text-xs text-gray-400 uppercase tracking-wide">Status</p>
            <div className="mt-1.5"><StatusBadge status={status.status} /></div>
          </Card>
        </div>
      </div>

      {/* Payment history this cycle */}
      <Card>
        <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-800">
          <div className="flex items-center gap-2">
            <Receipt size={18} className="text-gray-400" />
            <p className="text-sm font-semibold">Payment History — Current Cycle</p>
          </div>
          <span className="text-xs text-gray-400">{cyclePayments.length} payments</span>
        </div>
        {cyclePayments.length === 0 ? (
          <EmptyState
            icon={<Receipt size={32} />}
            title="No payments this cycle"
            message="No payments have been recorded for this tenant in the current cycle."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-gray-400 uppercase tracking-wide border-b border-gray-100 dark:border-gray-800">
                  <th className="text-left font-medium px-5 py-3">Date</th>
                  <th className="text-left font-medium px-5 py-3">Reference</th>
                  <th className="text-left font-medium px-5 py-3">Method</th>
                  <th className="text-right font-medium px-5 py-3">Amount</th>
                  <th className="text-right font-medium px-5 py-3">Receipt</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                {cyclePayments.map((p) => (
                  <tr key={p.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="px-5 py-3 text-gray-500">{formatDateShort(p.date)}</td>
                    <td className="px-5 py-3 font-mono text-xs text-gray-500">{p.reference}</td>
                    <td className="px-5 py-3 text-gray-500 capitalize">{p.method}</td>
                    <td className="px-5 py-3 text-right font-semibold text-emerald-600 dark:text-emerald-400">
                      +{formatCurrency(p.amount)}
                    </td>
                    <td className="px-5 py-3 text-right">
                      <button
                        className="btn btn-ghost p-1.5"
                        onClick={() => setReceiptData(buildReceiptData(p, payments))}
                        title="Print receipt"
                      >
                        <Printer size={15} />
                      </button>
                    </td>
                  </tr>
                ))}
                <tr className="border-t-2 border-gray-200 dark:border-gray-700">
                  <td className="px-5 py-3 font-semibold" colSpan={3}>Total Collected</td>
                  <td className="px-5 py-3 text-right font-bold">
                    {formatCurrency(status.totalPaidThisCycle)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <RecordPaymentModal
        open={showRecord}
        onClose={() => setShowRecord(false)}
        presetTenantId={tenantId}
        lockTenant
        recordedBy={account!.id}
      />
      <ReceiptModal
        open={receiptData !== null}
        onClose={() => setReceiptData(null)}
        data={receiptData}
      />
    </div>
  );
}
