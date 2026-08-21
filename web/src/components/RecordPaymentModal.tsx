import { useState } from 'react';
import type { Payment } from '../types';
import { useApp } from '../context';
import { Modal } from './Modal';
import { ReceiptModal, type ReceiptData } from './ReceiptModal';
import { buildReceiptData, formatCurrency, getTenant } from '../lib';
import { tenants } from '../data';
import { Banknote, Wallet, Smartphone, CreditCard, Check } from 'lucide-react';

const METHODS: { value: Payment['method']; label: string; icon: typeof Banknote }[] = [
  { value: 'bank', label: 'Bank Transfer', icon: Banknote },
  { value: 'mobile', label: 'Mobile Money', icon: Smartphone },
  { value: 'cash', label: 'Cash', icon: Wallet },
  { value: 'card', label: 'Card', icon: CreditCard },
];

export function RecordPaymentModal({
  open,
  onClose,
  presetTenantId,
  lockTenant,
  recordedBy,
}: {
  open: boolean;
  onClose: () => void;
  presetTenantId?: string;
  lockTenant?: boolean;
  recordedBy: string;
}) {
  const { recordPayment, payments } = useApp();
  const [tenantId, setTenantId] = useState(presetTenantId ?? '');
  const [amountStr, setAmountStr] = useState('');
  const [method, setMethod] = useState<Payment['method']>('bank');
  const [error, setError] = useState('');
  const [lastPayment, setLastPayment] = useState<Payment | null>(null);
  const [receiptData, setReceiptData] = useState<ReceiptData | null>(null);

  const selectedTenant = presetTenantId ?? tenantId;
  const tenant = selectedTenant ? getTenant(selectedTenant) : null;

  const handleSubmit = () => {
    const amount = parseFloat(amountStr);
    if (!selectedTenant) {
      setError('Select a tenant.');
      return;
    }
    if (!amount || amount <= 0) {
      setError('Enter a valid amount.');
      return;
    }
    recordPayment({ tenantId: selectedTenant, amount, method, recordedBy });
    const newPayment = payments[payments.length - 1];
    setLastPayment(newPayment);
    setAmountStr('');
    setTenantId('');
    setError('');
  };

  const handleClose = () => {
    setAmountStr('');
    setTenantId('');
    setError('');
    setLastPayment(null);
    onClose();
  };

  const handlePrintReceipt = () => {
    if (!lastPayment) return;
    setReceiptData(buildReceiptData(lastPayment, payments));
    setLastPayment(null);
    handleClose();
  };

  return (
    <>
      <Modal
        open={open}
        onClose={handleClose}
        title="Record Payment"
        footer={
          lastPayment ? (
            <>
              <button className="btn btn-secondary" onClick={handleClose}>
                Done
              </button>
              <button className="btn btn-primary" onClick={handlePrintReceipt}>
                Print Receipt
              </button>
            </>
          ) : (
            <>
              <button className="btn btn-secondary" onClick={handleClose}>
                Cancel
              </button>
              <button className="btn btn-primary" onClick={handleSubmit}>
                Record Payment
              </button>
            </>
          )
        }
      >
        {lastPayment ? (
          <div className="flex flex-col items-center text-center py-4">
            <div className="w-12 h-12 bg-emerald-50 text-emerald-600 dark:bg-emerald-950 dark:text-emerald-400 flex items-center justify-center mb-4">
              <Check size={24} />
            </div>
            <p className="font-semibold text-base">Payment Recorded</p>
            <p className="text-sm text-gray-500 dark:text-gray-400 mt-1">
              {formatCurrency(lastPayment.amount)} for {getTenant(lastPayment.tenantId)?.name}
            </p>
            <p className="text-xs text-gray-400 mt-2 font-mono">Ref: {lastPayment.reference}</p>
          </div>
        ) : (
          <div className="space-y-4">
            {error && (
              <div className="px-3 py-2 text-sm text-red-700 bg-red-50 border border-red-200 dark:text-red-300 dark:bg-red-950 dark:border-red-800">
                {error}
              </div>
            )}

            <div>
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Tenant
              </label>
              {lockTenant && tenant ? (
                <div className="input-base px-3 py-2 text-sm flex items-center justify-between">
                  <span>
                    {tenant.name} <span className="text-gray-400">· {tenant.email}</span>
                  </span>
                  <span className="text-xs text-gray-400">Locked</span>
                </div>
              ) : (
                <select
                  className="input-base w-full px-3 py-2 text-sm"
                  value={selectedTenant}
                  onChange={(e) => setTenantId(e.target.value)}
                >
                  <option value="">Select a tenant…</option>
                  {tenants.map((t) => (
                    <option key={t.id} value={t.id}>
                      {t.name} — {t.email}
                    </option>
                  ))}
                </select>
              )}
            </div>

            {tenant && (
              <div className="text-xs text-gray-500 dark:text-gray-400 flex items-center justify-between px-3 py-2 surface-muted">
                <span>Monthly Rent</span>
                <span className="font-semibold">{formatCurrency(tenant.monthlyRent)}</span>
              </div>
            )}

            <div>
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Amount (GHS)
              </label>
              <input
                type="number"
                className="input-base w-full px-3 py-2 text-sm"
                placeholder="0"
                value={amountStr}
                onChange={(e) => setAmountStr(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleSubmit()}
              />
            </div>

            <div>
              <label className="block text-xs font-semibold text-gray-500 dark:text-gray-400 mb-1.5 uppercase tracking-wide">
                Method
              </label>
              <div className="grid grid-cols-2 gap-2">
                {METHODS.map((m) => {
                  const active = method === m.value;
                  return (
                    <button
                      key={m.value}
                      type="button"
                      onClick={() => setMethod(m.value)}
                      className={`flex items-center gap-2 px-3 py-2 text-sm border transition-colors ${
                        active
                          ? 'border-brand-600 bg-brand-50 text-brand-700 dark:bg-brand-950 dark:text-brand-300 dark:border-brand-700'
                          : 'border-gray-300 text-gray-600 hover:bg-gray-50 dark:border-gray-700 dark:text-gray-300 dark:hover:bg-gray-800'
                      }`}
                    >
                      <m.icon size={16} />
                      {m.label}
                    </button>
                  );
                })}
              </div>
            </div>
          </div>
        )}
      </Modal>
      <ReceiptModal
        open={receiptData !== null}
        onClose={() => setReceiptData(null)}
        data={receiptData}
      />
    </>
  );
}
