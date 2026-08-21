import { useMemo, useState } from 'react';
import { ShoppingBag, Home, MapPin, ArrowLeft, Plus } from 'lucide-react';
import { useApp } from '../context';
import { Card, SectionHeader, StatusBadge } from './ui';
import { RecordPaymentModal } from './RecordPaymentModal';
import {
  buildAllLeaseStatuses,
  formatCurrency,
  formatDate,
  getPropertyTenants,
  getVacantUnits,
} from '../lib';
import { properties, units } from '../data';
import type { PropertyType } from '../types';

function PropertyIcon({ type, size = 18 }: { type: PropertyType; size?: number }) {
  return type === 'shopping' ? <ShoppingBag size={size} /> : <Home size={size} />;
}

function propertyTypeLabel(type: PropertyType): string {
  return type === 'shopping' ? 'Shopping Complex' : 'Apartment Complex';
}

export function PropertiesList({ onSelectProperty }: { onSelectProperty: (id: string) => void }) {
  const { payments } = useApp();
  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);
  // const vacant = getVacantUnits();

  return (
    <div>
      <SectionHeader
        title="Properties"
        subtitle={`${properties.length} properties across your portfolio`}
      />

      <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {properties.map((p) => {
          const propertyUnits = units.filter((u) => u.propertyId === p.id);
          const propertyStatuses = statuses.filter((s) => s.property.id === p.id);
          // const propertyVacant = vacant.filter((u) => u.propertyId === p.id);
          const totalRent = propertyStatuses.reduce((s, ls) => s + ls.monthlyRent, 0);
          const collected = propertyStatuses.reduce((s, ls) => s + ls.totalPaidThisCycle, 0);
          const occupancy = propertyUnits.length > 0
            ? Math.round((propertyStatuses.length / propertyUnits.length) * 100)
            : 0;

          return (
            <button
              key={p.id}
              onClick={() => onSelectProperty(p.id)}
              className="surface p-5 text-left hover:border-brand-400 dark:hover:border-brand-700 transition-colors group"
            >
              <div className="flex items-start justify-between mb-4">
                <div className="w-10 h-10 bg-brand-50 text-brand-600 dark:bg-brand-950 dark:text-brand-400 flex items-center justify-center">
                  <PropertyIcon type={p.type} />
                </div>
                <span className="text-xs font-medium text-gray-400 uppercase tracking-wide">
                  {propertyTypeLabel(p.type)}
                </span>
              </div>
              <p className="font-semibold text-base group-hover:text-brand-600 dark:group-hover:text-brand-400 transition-colors">
                {p.name}
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1 mt-1">
                <MapPin size={12} /> {p.address}
              </p>

              <div className="grid grid-cols-2 gap-3 mt-5 pt-4 border-t border-gray-100 dark:border-gray-800">
                <div>
                  <p className="text-xs text-gray-400">Units</p>
                  <p className="text-sm font-semibold">
                    {propertyStatuses.length}/{propertyUnits.length} occupied
                  </p>
                </div>
                <div>
                  <p className="text-xs text-gray-400">Occupancy</p>
                  <p className="text-sm font-semibold">{occupancy}%</p>
                </div>
                <div>
                  <p className="text-xs text-gray-400">Monthly Rent</p>
                  <p className="text-sm font-semibold">{formatCurrency(totalRent)}</p>
                </div>
                <div>
                  <p className="text-xs text-gray-400">Collected</p>
                  <p className="text-sm font-semibold text-emerald-600 dark:text-emerald-400">
                    {formatCurrency(collected)}
                  </p>
                </div>
              </div>
            </button>
          );
        })}
      </div>
    </div>
  );
}

export function PropertyDetail({
  propertyId,
  onBack,
  onSelectTenant,
}: {
  propertyId: string;
  onBack: () => void;
  onSelectTenant: (id: string) => void;
}) {
  const { account, payments } = useApp();
  const [showRecord, setShowRecord] = useState(false);

  const property = properties.find((p) => p.id === propertyId);
  const statuses = useMemo(() => buildAllLeaseStatuses(payments), [payments]);
  const propertyStatuses = statuses.filter((s) => s.property.id === propertyId);
  const propertyUnits = units.filter((u) => u.propertyId === propertyId);
  const vacant = getVacantUnits().filter((u) => u.propertyId === propertyId);
  const propTenants = getPropertyTenants(propertyId);

  if (!property) return null;

  const totalRent = propertyStatuses.reduce((s, ls) => s + ls.monthlyRent, 0);
  // const collected = propertyStatuses.reduce((s, ls) => s + ls.totalPaidThisCycle, 0);
  const outstanding = propertyStatuses.reduce((s, ls) => s + ls.balance, 0);

  return (
    <div>
      <button
        onClick={onBack}
        className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-gray-700 dark:hover:text-gray-300 mb-4"
      >
        <ArrowLeft size={16} /> Back to Properties
      </button>

      <SectionHeader
        title={property.name}
        subtitle={`${propertyTypeLabel(property.type)} · ${property.address}`}
        action={
          <button className="btn btn-primary" onClick={() => setShowRecord(true)}>
            <Plus size={16} /> Record Payment
          </button>
        }
      />

      {/* Summary stats */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 mb-6">
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Total Units</p>
          <p className="text-xl font-bold mt-1">{propertyUnits.length}</p>
          <p className="text-xs text-gray-400 mt-1">{vacant.length} vacant</p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Tenants</p>
          <p className="text-xl font-bold mt-1">{propTenants.length}</p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Monthly Rent</p>
          <p className="text-xl font-bold mt-1">{formatCurrency(totalRent)}</p>
        </Card>
        <Card className="p-4">
          <p className="text-xs text-gray-400 uppercase tracking-wide">Outstanding</p>
          <p className="text-xl font-bold mt-1 text-amber-600 dark:text-amber-400">
            {formatCurrency(outstanding)}
          </p>
        </Card>
      </div>

      {/* Units table */}
      <Card>
        <div className="px-5 py-4 border-b border-gray-200 dark:border-gray-800">
          <p className="text-sm font-semibold">Units & Tenants</p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="text-xs text-gray-400 uppercase tracking-wide border-b border-gray-100 dark:border-gray-800">
                <th className="text-left font-medium px-5 py-3">Unit</th>
                <th className="text-left font-medium px-5 py-3">Floor</th>
                <th className="text-left font-medium px-5 py-3">Tenant</th>
                <th className="text-left font-medium px-5 py-3">Rent</th>
                <th className="text-left font-medium px-5 py-3">Paid</th>
                <th className="text-left font-medium px-5 py-3">Balance</th>
                <th className="text-left font-medium px-5 py-3">Due</th>
                <th className="text-left font-medium px-5 py-3">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100 dark:divide-gray-800">
              {propertyUnits.map((u) => {
                const status = propertyStatuses.find((s) => s.unit.id === u.id);
                if (!status) {
                  return (
                    <tr key={u.id} className="hover:bg-gray-50 dark:hover:bg-gray-800/50">
                      <td className="px-5 py-3 font-medium">{u.label}</td>
                      <td className="px-5 py-3 text-gray-400">F{u.floor}</td>
                      <td className="px-5 py-3 text-gray-400">—</td>
                      <td className="px-5 py-3 text-gray-400">—</td>
                      <td className="px-5 py-3 text-gray-400">—</td>
                      <td className="px-5 py-3 text-gray-400">—</td>
                      <td className="px-5 py-3 text-gray-400">—</td>
                      <td className="px-5 py-3"><StatusBadge status="vacant" /></td>
                    </tr>
                  );
                }
                return (
                  <tr
                    key={u.id}
                    className="hover:bg-gray-50 dark:hover:bg-gray-800/50 cursor-pointer"
                    onClick={() => onSelectTenant(status.tenant.id)}
                  >
                    <td className="px-5 py-3 font-medium">{u.label}</td>
                    <td className="px-5 py-3 text-gray-400">F{u.floor}</td>
                    <td className="px-5 py-3">{status.tenant.name}</td>
                    <td className="px-5 py-3">{formatCurrency(status.monthlyRent)}</td>
                    <td className="px-5 py-3 text-emerald-600 dark:text-emerald-400">
                      {formatCurrency(status.totalPaidThisCycle)}
                    </td>
                    <td className="px-5 py-3">
                      {status.balance > 0 ? formatCurrency(status.balance) : '—'}
                    </td>
                    <td className="px-5 py-3 text-gray-500">{formatDate(status.nextDueDate)}</td>
                    <td className="px-5 py-3"><StatusBadge status={status.status} /></td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      </Card>

      <RecordPaymentModal
        open={showRecord}
        onClose={() => setShowRecord(false)}
        recordedBy={account!.id}
      />
    </div>
  );
}
