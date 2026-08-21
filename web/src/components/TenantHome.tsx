import { useMemo, useState } from 'react';
import {
  Home,
  Calendar,
  Wallet,
  Receipt,
  Plus,
  CheckCircle2,
  AlertTriangle,
  Clock,
  TrendingUp,
} from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, StatusBadge, ProgressBar, EmptyState } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import { ReceiptModal, type ReceiptData } from './ReceiptModal';
import { Printer } from 'lucide-react';
import {
  buildLeaseStatus,
  buildReceiptData,
  formatCurrency,
  formatDate,
  formatDateShort,
  getCyclePayments,
} from '../lib';
import { tenants, units, properties } from '../data';

export function TenantHome() {
  const { account, payments } = useApp();
  const [showPay, setShowPay] = useState(false);
  const [receiptData, setReceiptData] = useState<ReceiptData | null>(null);

  const tenant = useMemo(
    () => tenants.find((t) => t.id === account?.tenantId),
    [account],
  );

  const status = useMemo(
    () => (tenant ? buildLeaseStatus(tenant, payments, units, properties) : null),
    [tenant, payments],
  );

  const statusFull = status;

  const cyclePayments = useMemo(
    () => (tenant ? getCyclePayments(tenant.id, payments) : []),
    [tenant, payments],
  );

  if (!tenant || !statusFull) {
    return (
      <div>
        <SectionHeader title="My Account" />
        <EmptyState
          icon={<Home size={32} />}
          title="No tenant profile"
          message="Your demo account is not linked to a tenant record."
        />
      </div>
    );
  }

  const pct = (statusFull.totalPaidThisCycle / statusFull.monthlyRent) * 100;
  const isPaid = statusFull.status === 'paid';
  const isOverdue = statusFull.status === 'overdue';

  return (
    <div>
      <SectionHeader
        title={`Welcome, ${tenant.name.split(' ')[0]}`}
        subtitle="Your lease and payment information at a glance"
        action={
          <button className="btn btn-primary" onClick={() => setShowPay(true)}>
            <Plus size={16} /> Make a Payment
          </button>
        }
      />

      {/* Status banner */}
      <div
        className={`surface p-5 mb-6 border-l-4 ${
          isPaid
            ? 'border-l-emerald-500'
            : isOverdue
              ? 'border-l-red-500'
              : 'border-l-amber-500'
        }`}
      >
        <div className="flex items-center gap-3">
          {isPaid ? (
            <CheckCircle2 className="text-emerald-500" size={24} />
          ) : isOverdue ? (
            <AlertTriangle className="text-red-500" size={24} />
          ) : (
            <Clock className="text-amber-500" size={24} />
          )}
          <div className="flex-1">
            <p className="font-semibold">
              {isPaid
                ? 'Your rent is fully paid this cycle'
                : isOverdue
                  ? `Your rent is overdue by ${Math.abs(statusFull.daysUntilDue)} days`
                  : `Partial payment — ${formatCurrency(statusFull.balance)} remaining`}
            </p>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              Next due date: {formatDate(statusFull.nextDueDate)}
            </p>
          </div>
          <StatusBadge status={statusFull.status} />
        </div>
      </div>

      {/* Key stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <Home size={14} /> Property
          </div>
          <p className="text-sm font-semibold mt-2">{statusFull.property.name}</p>
          <p className="text-xs text-gray-400 mt-0.5">{statusFull.unit.label}</p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <Wallet size={14} /> Monthly Rent
          </div>
          <p className="text-xl font-bold mt-2">{formatCurrency(statusFull.monthlyRent)}</p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <TrendingUp size={14} /> Paid This Cycle
          </div>
          <p className="text-xl font-bold mt-2 text-emerald-600 dark:text-emerald-400">
            {formatCurrency(statusFull.totalPaidThisCycle)}
          </p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <AlertTriangle size={14} /> Balance Due
          </div>
          <p className="text-xl font-bold mt-2 text-amber-600 dark:text-amber-400">
            {formatCurrency(statusFull.balance)}
          </p>
        </Card>
      </div>

      {/* Lease details + progress */}
      <div className="grid lg:grid-cols-3 gap-6 mb-6">
        <Card className="p-5 lg:col-span-2">
          <p className="text-sm font-semibold mb-4">Lease Details</p>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Property</p>
              <p className="text-sm font-medium mt-1">{statusFull.property.name}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Unit</p>
              <p className="text-sm font-medium mt-1">{statusFull.unit.label}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Lease Start</p>
              <p className="text-sm font-medium mt-1">{formatDate(tenant.leaseStart)}</p>
            </div>
            <div>
              <p className="text-xs text-gray-400 uppercase tracking-wide">Due Day</p>
              <p className="text-sm font-medium mt-1">{tenant.dueDayOfMonth}th each month</p>
            </div>
          </div>
          <div className="pt-4 mt-4 border-t border-gray-100 dark:border-gray-800">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-gray-500 dark:text-gray-400">
                {formatCurrency(statusFull.totalPaidThisCycle)} of {formatCurrency(statusFull.monthlyRent)} paid
              </span>
              <span className="text-sm font-bold">{pct.toFixed(0)}%</span>
            </div>
            <ProgressBar value={statusFull.totalPaidThisCycle} max={statusFull.monthlyRent} />
          </div>
        </Card>

        <Card className="p-5">
          <div className="flex items-center gap-2 mb-4">
            <Calendar size={18} className="text-gray-400" />
            <p className="text-sm font-semibold">Next Due Date</p>
          </div>
          <p className="text-2xl font-bold">{formatDate(statusFull.nextDueDate)}</p>
          <p
            className={`text-sm mt-1 ${
              statusFull.daysUntilDue < 0
                ? 'text-red-500'
                : statusFull.daysUntilDue <= 3
                  ? 'text-amber-500'
                  : 'text-gray-400'
            }`}
          >
            {statusFull.daysUntilDue < 0
              ? `${Math.abs(statusFull.daysUntilDue)} days overdue`
              : statusFull.daysUntilDue === 0
                ? 'Due today'
                : `${statusFull.daysUntilDue} days from now`}
          </p>
          <div className="mt-4 pt-4 border-t border-gray-100 dark:border-gray-800">
            <p className="text-xs text-gray-400 uppercase tracking-wide mb-1">Amount Due</p>
            <p className="text-lg font-bold">{formatCurrency(statusFull.balance)}</p>
          </div>
        </Card>
      </div>

      {/* Payment history */}
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
            title="No payments yet"
            message="You have not made any payments this cycle. Click 'Make a Payment' to get started."
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
                  <td className="px-5 py-3 font-semibold" colSpan={3}>Total Paid</td>
                  <td className="px-5 py-3 text-right font-bold">
                    {formatCurrency(statusFull.totalPaidThisCycle)}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        )}
      </Card>

      <RecordPaymentModal
        open={showPay}
        onClose={() => setShowPay(false)}
        presetTenantId={tenant.id}
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
