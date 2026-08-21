export type Role = 'owner' | 'manager' | 'tenant';

export type PropertyType = 'shopping' | 'apartment';

export interface Property {
  id: string;
  name: string;
  type: PropertyType;
  address: string;
  managerId: string | null;
}

export interface Unit {
  id: string;
  propertyId: string;
  label: string;
  floor: number;
}

export interface Tenant {
  id: string;
  name: string;
  email: string;
  phone: string;
  unitId: string;
  propertyId: string;
  leaseStart: string;
  monthlyRent: number;
  dueDayOfMonth: number;
}

export interface Payment {
  id: string;
  tenantId: string;
  amount: number;
  date: string;
  method: 'bank' | 'cash' | 'mobile' | 'card';
  reference: string;
  recordedBy: string;
}

export interface DemoAccount {
  id: string;
  role: Role;
  name: string;
  email: string;
  phone: string;
  password: string;
  tenantId?: string;
  managerId?: string;
}

export type PaymentStatus = 'paid' | 'partial' | 'overdue' | 'vacant';

export interface LeaseStatus {
  tenant: Tenant;
  unit: Unit;
  property: Property;
  monthlyRent: number;
  totalPaidThisCycle: number;
  balance: number;
  status: PaymentStatus;
  nextDueDate: string;
  daysUntilDue: number;
}
