import { useMemo, useState } from 'react';
import {
  Building2,
  Wallet,
  AlertTriangle,
  TrendingUp,
  CalendarClock,
  Receipt,
  Plus,
  ArrowRight,
} from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, ProgressBar } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import {
  buildAllLeaseStatuses,
  formatCurrency,
  formatDateShort,
  formatCurrency as fmt,
  getVacantUnits,
} from '../lib';
import { properties } from '../data';
import type { Page } from './Layout';

function StatCard({
  label,
  value,
  sub,
  icon: Icon,
  tone = 'default',
}: {
  label: string;
  value: string;
  sub?: string;
  icon: typeof Building2;
  tone?: 'default' | 'success' | 'warning' | 'danger';
}) {
  const toneCls = {
    default: 'text-brand-600 dark:text-brand-400 bg-brand-50 dark:bg-brand-950',
    success: 'text-emerald-600 dark:text-emerald-400 bg-emerald-50 dark:bg-emerald-950',
    warning: 'text-amber-600 dark:text-amber-400 bg-amber-50 dark:bg-amber-950',
    danger: 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-950',
  }[tone];

  return (
    <Card className="p-5">
      <div className="flex items-start justify-between">
        <div className={`w-10 h-10 flex items-center justify-center ${toneCls}`}>
          <Icon size={20} />
        </div>
      </div>
      <p className="text-2xl font-bold mt-4 tracking-tight">{value}</p>
      <p className="text-xs text-gray-500 dark:text-gray-400 mt-1">{label}</p>
      {sub && <p className="text-xs text-gray-400 dark:text-gray-500 mt-2">{sub}</p>}
    </Card>
  );
}

export function Overview({ onNavigate }: { onNavigate: (p: Page) => void }) {
  const { account, payments } = useApp();
  const [showRecord, setShowRecord] = useState(false);

  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);
  const vacant = getVacantUnits();

  const totalExpected = statuses.reduce((s, ls) => s + ls.monthlyRent, 0);
  const totalCollected = statuses.reduce((s, ls) => s + ls.totalPaidThisCycle, 0);
  const outstanding = statuses.reduce((s, ls) => s + ls.balance, 0);
  const overdueCount = statuses.filter((s) => s.status === 'overdue').length;
  const partialCount = statuses.filter((s) => s.status === 'partial').length;
  const paidCount = statuses.filter((s) => s.status === 'paid').length;
  const collectionRate = totalExpected > 0 ? (totalCollected / totalExpected) * 100 : 0;

  const upcoming = useMemo(
    () =>
      [...statuses]
        .filter((s) => s.status !== 'paid')
        .sort((a, b) => a.daysUntilDue - b.daysUntilDue)
        .slice(0, 6),
    [statuses],
  );

  const recentPayments = useMemo(
    () =>
      [...payments].sort((a, b) => b.date.localeCompare(a.date)).slice(0, 6),
    [payments],
  );

  const tenantName = (id: string) => statuses.find((s) => s.tenant.id === id)?.tenant.name ?? '—';

  return (
    <div>
      <SectionHeader
        title="Overview"
        subtitle={`Rent cycle summary for ${properties.length} properties`}
        action={
          <button className="btn btn-primary" onClick={() => setShowRecord(true)}>
            <Plus size={16} /> Record Payment
          </button>
        }
      />

      {/* Stats grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <StatCard
          label="Total Monthly Rent"
          value={formatCurrency(totalExpected)}
          sub={`${statuses.length} active leases · ${vacant.length} vacant`}
          icon={Wallet}
        />
        <StatCard
          label="Collected This Cycle"
          value={formatCurrency(totalCollected)}
          sub={`${collectionRate.toFixed(0)}% of expected`}
          icon={TrendingUp}
          tone="success"
        />
        <StatCard
          label="Outstanding Balance"
          value={formatCurrency(outstanding)}
          sub={`${partialCount} partial · ${overdueCount} overdue`}
          icon={AlertTriangle}
          tone={outstanding > 0 ? 'warning' : 'default'}
        />
        <StatCard
          label="Overdue Tenants"
          value={String(overdueCount)}
          sub={overdueCount > 0 ? 'Needs attention' : 'All on track'}
          icon={AlertTriangle}
          tone={overdueCount > 0 ? 'danger' : 'default'}
        />
      </div>

      {/* Collection progress */}
      <Card className="p-5 mb-6">
        <div className="flex items-center justify-between mb-3">
          <div>
            <p className="text-sm font-semibold">Rent Collection Progress</p>
            <p className="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
              {formatCurrency(totalCollected)} of {formatCurrency(totalExpected)} collected
            </p>
          </div>
          <span className="text-2xl font-bold text-brand-600 dark:text-brand-400">
            {collectionRate.toFixed(0)}%
          </span>
        </div>
        <ProgressBar value={totalCollected} max={totalExpected} />
        <div className="flex items-center gap-5 mt-4 text-xs">
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 bg-emerald-500" /> Paid ({paidCount})
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 bg-amber-500" /> Partial ({partialCount})
          </span>
          <span className="flex items-center gap-1.5">
            <span className="w-2.5 h-2.5 bg-red-500" /> Overdue ({overdueCount})
          </span>
        </div>
      </Card>

      <div className="grid lg:grid-cols-2 gap-6">
        {/* Upcoming due dates */}
        <Card>
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-800">
            <div className="flex items-center gap-2">
              <CalendarClock size={18} className="text-gray-400" />
              <p className="text-sm font-semibold">Upcoming & Overdue</p>
            </div>
            <button
              className="text-xs text-brand-600 dark:text-brand-400 font-medium hover:underline flex items-center gap-1"
              onClick={() => onNavigate('due-dates')}
            >
              View all <ArrowRight size={12} />
            </button>
          </div>
          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {upcoming.length === 0 && (
              <p className="px-5 py-8 text-sm text-gray-400 text-center">All rents collected this cycle.</p>
            )}
            {upcoming.map((s) => (
              <div key={s.tenant.id} className="px-5 py-3 flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{s.tenant.name}</p>
                  <p className="text-xs text-gray-400 truncate">
                    {s.property.name} · {s.unit.label}
                  </p>
                </div>
                <div className="text-right shrink-0">
                  <p className="text-sm font-semibold">{fmt(s.balance)}</p>
                  <p
                    className={`text-xs ${
                      s.daysUntilDue < 0
                        ? 'text-red-500'
                        : s.daysUntilDue <= 3
                          ? 'text-amber-500'
                          : 'text-gray-400'
                    }`}
                  >
                    {s.daysUntilDue < 0
                      ? `${Math.abs(s.daysUntilDue)}d overdue`
                      : s.daysUntilDue === 0
                        ? 'Due today'
                        : `in ${s.daysUntilDue}d`}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </Card>

        {/* Recent payments */}
        <Card>
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-800">
            <div className="flex items-center gap-2">
              <Receipt size={18} className="text-gray-400" />
              <p className="text-sm font-semibold">Recent Payments</p>
            </div>
            <button
              className="text-xs text-brand-600 dark:text-brand-400 font-medium hover:underline flex items-center gap-1"
              onClick={() => onNavigate('payments')}
            >
              View all <ArrowRight size={12} />
            </button>
          </div>
          <div className="divide-y divide-gray-100 dark:divide-gray-800">
            {recentPayments.map((p) => (
              <div key={p.id} className="px-5 py-3 flex items-center justify-between gap-3">
                <div className="min-w-0">
                  <p className="text-sm font-medium truncate">{tenantName(p.tenantId)}</p>
                  <p className="text-xs text-gray-400">
                    {formatDateShort(p.date)} · {p.method} · {p.reference}
                  </p>
                </div>
                <p className="text-sm font-semibold text-emerald-600 dark:text-emerald-400 shrink-0">
                  +{fmt(p.amount)}
                </p>
              </div>
            ))}
          </div>
        </Card>
      </div>

      <RecordPaymentModal
        open={showRecord}
        onClose={() => setShowRecord(false)}
        recordedBy={account!.id}
      />
    </div>
  );
}
