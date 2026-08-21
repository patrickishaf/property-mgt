import { useEffect } from 'react';
import { X, Printer, Building, Check } from 'lucide-react';
import type { Payment, Tenant, Property, Unit } from '../types';
import { formatCurrency, formatDate } from '../lib';

export interface ReceiptData {
  payment: Payment;
  tenant: Tenant;
  property: Property;
  unit: Unit;
  monthlyRent: number;
  totalPaidThisCycle: number;
  balance: number;
  paymentType: 'full' | 'partial';
  paymentNumber: number;
}

export function ReceiptModal({
  open,
  onClose,
  data,
}: {
  open: boolean;
  onClose: () => void;
  data: ReceiptData | null;
}) {
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', handler);
    return () => window.removeEventListener('keydown', handler);
  }, [open, onClose]);

  if (!open || !data) return null;

  // const { payment, tenant, property, unit, monthlyRent, totalPaidThisCycle, balance, paymentType, paymentNumber } = data;

  const handlePrint = () => {
    document.body.classList.add('printing-receipt');
    window.print();
    document.body.classList.remove('printing-receipt');
  };

  return (
    <>
      {/* On-screen modal */}
      <div className="fixed inset-0 z-50 flex items-center justify-center p-4 print:hidden">
        <div className="absolute inset-0 bg-black/40" onClick={onClose} />
        <div className="relative w-full max-w-md surface shadow-xl max-h-[90vh] overflow-y-auto">
          <div className="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-800">
            <h2 className="text-base font-semibold">Payment Receipt</h2>
            <button onClick={onClose} className="btn-ghost p-1 -mr-1" aria-label="Close">
              <X size={18} />
            </button>
          </div>

          {/* Receipt content — shared with print version */}
          <div className="px-5 py-4">
            <ReceiptContent data={data} />
          </div>

          <div className="flex items-center justify-end gap-2 px-5 py-4 border-t border-gray-200 dark:border-gray-800">
            <button className="btn btn-secondary" onClick={onClose}>
              Close
            </button>
            <button className="btn btn-primary" onClick={handlePrint}>
              <Printer size={16} /> Print Receipt
            </button>
          </div>
        </div>
      </div>

      {/* Print-only version */}
      <div className="hidden print:block printable-receipt">
        <ReceiptContent data={data} />
      </div>
    </>
  );
}

function ReceiptContent({ data }: { data: ReceiptData }) {
  const { payment, tenant, property, unit, monthlyRent, totalPaidThisCycle, balance, paymentType, paymentNumber } = data;

  return (
    <div className="receipt-body">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-gray-300 pb-4 mb-4">
        <div className="flex items-center gap-2.5">
          <div className="w-9 h-9 bg-brand-600 flex items-center justify-center">
            <Building size={18} className="text-white" />
          </div>
          <div>
            <p className="font-bold text-base leading-tight">RentLedger</p>
            <p className="text-[10px] text-gray-500 uppercase tracking-wider">Property Management</p>
          </div>
        </div>
        <div className="text-right">
          <p className="text-xs font-semibold text-gray-500 uppercase tracking-wide">Receipt</p>
          <p className="font-mono text-sm font-semibold">{payment.reference}</p>
        </div>
      </div>

      {/* Payment type banner */}
      <div className="flex items-center gap-2 mb-4 border border-gray-200 px-3 py-2">
        <Check size={16} className="text-emerald-600" />
        <span className="text-sm font-semibold">
          {paymentType === 'full' ? 'Full Payment' : 'Part Payment'} — Payment #{paymentNumber}
        </span>
      </div>

      {/* Tenant & property info */}
      <div className="grid grid-cols-2 gap-x-4 gap-y-3 mb-4 text-sm">
        <div>
          <p className="text-[10px] text-gray-400 uppercase tracking-wide">Received From</p>
          <p className="font-semibold">{tenant.name}</p>
          <p className="text-xs text-gray-500">{tenant.email}</p>
          <p className="text-xs text-gray-500">{tenant.phone}</p>
        </div>
        <div>
          <p className="text-[10px] text-gray-400 uppercase tracking-wide">Property</p>
          <p className="font-semibold">{property.name}</p>
          <p className="text-xs text-gray-500">{property.address}</p>
          <p className="text-xs text-gray-500">Unit: {unit.label}</p>
        </div>
      </div>

      {/* Payment details table */}
      <table className="w-full text-sm border-t border-gray-300">
        <tbody>
          <tr className="border-b border-gray-200">
            <td className="py-2 text-gray-500">Payment Date</td>
            <td className="py-2 text-right font-medium">{formatDate(payment.date)}</td>
          </tr>
          <tr className="border-b border-gray-200">
            <td className="py-2 text-gray-500">Payment Method</td>
            <td className="py-2 text-right font-medium capitalize">{payment.method}</td>
          </tr>
          <tr className="border-b border-gray-200">
            <td className="py-2 text-gray-500">Reference No.</td>
            <td className="py-2 text-right font-mono text-xs font-medium">{payment.reference}</td>
          </tr>
          <tr className="border-b border-gray-200">
            <td className="py-2 text-gray-500">Monthly Rent</td>
            <td className="py-2 text-right font-medium">{formatCurrency(monthlyRent)}</td>
          </tr>
          <tr className="border-b-2 border-gray-300">
            <td className="py-2 text-gray-500">Total Paid This Cycle</td>
            <td className="py-2 text-right font-medium">{formatCurrency(totalPaidThisCycle)}</td>
          </tr>
          <tr className="border-b border-gray-200">
            <td className="py-2 font-semibold">Amount Paid</td>
            <td className="py-2 text-right text-lg font-bold text-emerald-600">
              {formatCurrency(payment.amount)}
            </td>
          </tr>
          <tr className="border-b border-gray-200">
            <td className="py-2 text-gray-500">Remaining Balance</td>
            <td className="py-2 text-right font-medium">
              {balance > 0 ? formatCurrency(balance) : 'GHS 0 — Fully Paid'}
            </td>
          </tr>
        </tbody>
      </table>

      {/* Footer */}
      <div className="mt-6 pt-4 border-t border-gray-300 text-center">
        <p className="text-xs text-gray-400">
          This is a computer-generated receipt and does not require a signature.
        </p>
        <p className="text-[10px] text-gray-400 mt-1">
          RentLedger Property Management · Generated on {formatDate(new Date().toISOString().slice(0, 10))}
        </p>
      </div>
    </div>
  );
}
