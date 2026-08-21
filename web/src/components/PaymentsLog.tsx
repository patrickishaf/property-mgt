import { useMemo, useState } from 'react';
import { Search, Plus, Receipt, Banknote, Smartphone, Wallet, CreditCard } from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, EmptyState } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import { ReceiptModal, type ReceiptData } from './ReceiptModal';
import { buildReceiptData } from '../lib';
import { Printer } from 'lucide-react';
import { formatCurrency, formatDateShort, buildAllLeaseStatuses } from '../lib';
import { properties } from '../data';
import type { Payment } from '../types';

const METHOD_ICON: Record<Payment['method'], typeof Banknote> = {
  bank: Banknote,
  mobile: Smartphone,
  cash: Wallet,
  card: CreditCard,
};

export function PaymentsLog() {
  const { account, payments } = useApp();
  const [showRecord, setShowRecord] = useState(false);
  const [receiptData, setReceiptData] = useState<ReceiptData | null>(null);
  const [search, setSearch] = useState('');
  const [methodFilter, setMethodFilter] = useState<string>('all');
  const [propertyFilter, setPropertyFilter] = useState<string>('all');

  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);
  const tenantMap = useMemo(() => {
    const m = new Map<string, { name: string; property: string; unit: string }>();
    statuses.forEach((s) => {
      m.set(s.tenant.id, {
        name: s.tenant.name,
        property: s.property.name,
        unit: s.unit.label,
      });
    });
    return m;
  }, [statuses]);

  const filtered = useMemo(() => {
    return [...payments]
      .sort((a, b) => b.date.localeCompare(a.date))
      .filter((p) => {
        const info = tenantMap.get(p.tenantId);
        const matchesSearch =
          !search ||
          (info?.name.toLowerCase().includes(search.toLowerCase()) ?? false) ||
          p.reference.toLowerCase().includes(search.toLowerCase());
        const matchesMethod = methodFilter === 'all' || p.method === methodFilter;
        const matchesProperty =
          propertyFilter === 'all' ||
          statuses.find((s) => s.tenant.id === p.tenantId)?.property.id === propertyFilter;
        return matchesSearch && matchesMethod && matchesProperty;
      });
  }, [payments, search, methodFilter, propertyFilter, tenantMap, statuses]);

  const totalAmount = filtered.reduce((s, p) => s + p.amount, 0);

  return (
    <div>
      <SectionHeader
        title="Payments"
        subtitle={`${payments.length} total payments recorded`}
        action={
          <button className="btn btn-primary" onClick={() => setShowRecord(true)}>
            <Plus size={16} /> Record Payment
          </button>
        }
      />

      {/* Summary */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Total Collected</p>
          <p className="text-xl font-bold mt-1">{formatCurrency(totalAmount)}</p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Filtered</p>
          <p className="text-xl font-bold mt-1">{filtered.length}</p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">This Cycle</p>
          <p className="text-xl font-bold mt-1">
            {formatCurrency(
              payments
                .filter((p) => p.date >= '2025-08-01')
                .reduce((s, p) => s + p.amount, 0),
            )}
          </p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">All Time</p>
          <p className="text-xl font-bold mt-1">{formatCurrency(payments.reduce((s, p) => s + p.amount, 0))}</p>
        </Card>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-5">
        <div className="relative flex-1">
          <Search size={16} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" />
          <input
            type="text"
            placeholder="Search by tenant or reference…"
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
          value={methodFilter}
          onChange={(e) => setMethodFilter(e.target.value)}
        >
          <option value="all">All Methods</option>
          <option value="bank">Bank Transfer</option>
          <option value="mobile">Mobile Money</option>
          <option value="cash">Cash</option>
          <option value="card">Card</option>
        </select>
      </div>

      {/* Table */}
      <Card>
        {filtered.length === 0 ? (
          <EmptyState
            icon={<Receipt size={32} />}
            title="No payments found"
            message="Try adjusting your search or filters."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-gray-400 uppercase tracking-wide border-b border-gray-100 dark:border-gray-800">
                  <th className="text-left font-medium px-5 py-3">Date</th>
                  <th className="text-left font-medium px-5 py-3">Tenant</th>
                  <th className="text-left font-medium px-5 py-3">Property</th>
                  <th className="text-left font-medium px-5 py-3">Unit</th>
                  <th className="text-left font-medium px-5 py-3">Method</th>
                  <th className="text-left font-medium px-5 py-3">Reference</th>
                  <th className="text-right font-medium px-5 py-3">Amount</th>
                  <th className="text-right font-medium px-5 py-3">Receipt</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                {filtered.map((p) => {
                  const info = tenantMap.get(p.tenantId);
                  const Icon = METHOD_ICON[p.method];
                  return (
                    <tr key={p.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                      <td className="px-5 py-3 text-gray-500 whitespace-nowrap">{formatDateShort(p.date)}</td>
                      <td className="px-5 py-3 font-medium">{info?.name ?? '—'}</td>
                      <td className="px-5 py-3 text-gray-500">{info?.property ?? '—'}</td>
                      <td className="px-5 py-3 text-gray-500">{info?.unit ?? '—'}</td>
                      <td className="px-5 py-3">
                        <span className="flex items-center gap-1.5 text-gray-500 capitalize">
                          <Icon size={14} /> {p.method}
                        </span>
                      </td>
                      <td className="px-5 py-3 font-mono text-xs text-gray-400">{p.reference}</td>
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
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <RecordPaymentModal
        open={showRecord}
        onClose={() => setShowRecord(false)}
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
