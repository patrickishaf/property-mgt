import { useMemo, useState } from 'react';
import { CalendarClock, AlertTriangle, Clock, CheckCircle2, Plus } from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, StatusBadge, EmptyState } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import { buildAllLeaseStatuses, formatCurrency, formatDate } from '../lib';

type Tab = 'all' | 'overdue' | 'upcoming' | 'paid';

export function DueDates() {
  const { account, payments } = useApp();
  const [showRecord, setShowRecord] = useState(false);
  const [tab, setTab] = useState<Tab>('all');

  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);

  const overdue = statuses.filter((s) => s.status === 'overdue');
  const partial = statuses.filter((s) => s.status === 'partial' && s.daysUntilDue >= 0);
  const paid = statuses.filter((s) => s.status === 'paid');

  const sorted = useMemo(() => {
    const list = [...statuses];
    switch (tab) {
      case 'overdue':
        return list.filter((s) => s.status === 'overdue').sort((a, b) => a.daysUntilDue - b.daysUntilDue);
      case 'upcoming':
        return list
          .filter((s) => s.status === 'partial' && s.daysUntilDue >= 0)
          .sort((a, b) => a.daysUntilDue - b.daysUntilDue);
      case 'paid':
        return list.filter((s) => s.status === 'paid');
      default:
        return list.sort((a, b) => a.daysUntilDue - b.daysUntilDue);
    }
  }, [statuses, tab]);

  const tabs: { id: Tab; label: string; count: number; icon: typeof Clock; tone: string }[] = [
    { id: 'all', label: 'All', count: statuses.length, icon: CalendarClock, tone: 'text-gray-500' },
    { id: 'overdue', label: 'Overdue', count: overdue.length, icon: AlertTriangle, tone: 'text-red-500' },
    { id: 'upcoming', label: 'Upcoming', count: partial.length, icon: Clock, tone: 'text-amber-500' },
    { id: 'paid', label: 'Paid', count: paid.length, icon: CheckCircle2, tone: 'text-emerald-500' },
  ];

  return (
    <div>
      <SectionHeader
        title="Due Dates"
        subtitle="Track rent due dates and outstanding balances across all tenants"
        action={
          <button className="btn btn-primary" onClick={() => setShowRecord(true)}>
            <Plus size={16} /> Record Payment
          </button>
        }
      />

      {/* Summary cards */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <CalendarClock size={14} /> Total Leases
          </div>
          <p className="text-xl font-bold mt-2">{statuses.length}</p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <AlertTriangle size={14} /> Overdue
          </div>
          <p className="text-xl font-bold mt-2 text-red-500">{overdue.length}</p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <Clock size={14} /> Partial / Upcoming
          </div>
          <p className="text-xl font-bold mt-2 text-amber-500">{partial.length}</p>
        </Card>
        <Card className="p-4">
          <div className="flex items-center gap-2 text-xs text-gray-400 uppercase tracking-wide">
            <CheckCircle2 size={14} /> Paid
          </div>
          <p className="text-xl font-bold mt-2 text-emerald-500">{paid.length}</p>
        </Card>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 mb-5 border-b border-gray-200 dark:border-gray-800">
        {tabs.map((t) => (
          <button
            key={t.id}
            onClick={() => setTab(t.id)}
            className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-colors -mb-px ${
              tab === t.id
                ? 'border-brand-600 text-brand-600 dark:text-brand-400 dark:border-brand-500'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
            }`}
          >
            <t.icon size={15} className={tab === t.id ? '' : t.tone} />
            {t.label}
            <span className="text-xs px-1.5 py-0.5 bg-gray-100 dark:bg-gray-800 text-gray-500">
              {t.count}
            </span>
          </button>
        ))}
      </div>

      {/* Table */}
      <Card>
        {sorted.length === 0 ? (
          <EmptyState
            icon={<CalendarClock size={32} />}
            title="Nothing here"
            message="No tenants match this filter."
          />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-xs text-gray-400 uppercase tracking-wide border-b border-gray-100 dark:border-gray-800">
                  <th className="text-left font-medium px-5 py-3">Tenant</th>
                  <th className="text-left font-medium px-5 py-3">Property</th>
                  <th className="text-left font-medium px-5 py-3">Unit</th>
                  <th className="text-right font-medium px-5 py-3">Rent</th>
                  <th className="text-right font-medium px-5 py-3">Paid</th>
                  <th className="text-right font-medium px-5 py-3">Balance</th>
                  <th className="text-left font-medium px-5 py-3">Due Date</th>
                  <th className="text-left font-medium px-5 py-3">Days</th>
                  <th className="text-left font-medium px-5 py-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
                {sorted.map((s) => (
                  <tr key={s.tenant.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                    <td className="px-5 py-3 font-medium">{s.tenant.name}</td>
                    <td className="px-5 py-3 text-gray-500">{s.property.name}</td>
                    <td className="px-5 py-3 text-gray-500">{s.unit.label}</td>
                    <td className="px-5 py-3 text-right">{formatCurrency(s.monthlyRent)}</td>
                    <td className="px-5 py-3 text-right text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(s.totalPaidThisCycle)}
                    </td>
                    <td className="px-5 py-3 text-right">
                      {s.balance > 0 ? (
                        <span className="text-amber-600 dark:text-amber-400 font-medium">
                          {formatCurrency(s.balance)}
                        </span>
                      ) : (
                        <span className="text-gray-400">—</span>
                      )}
                    </td>
                    <td className="px-5 py-3 text-gray-500 whitespace-nowrap">{formatDate(s.nextDueDate)}</td>
                    <td className="px-5 py-3">
                      <span
                        className={`text-xs font-medium ${
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
                            ? 'Today'
                            : `in ${s.daysUntilDue}d`}
                      </span>
                    </td>
                    <td className="px-5 py-3"><StatusBadge status={s.status} /></td>
                  </tr>
                ))}
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
    </div>
  );
}
