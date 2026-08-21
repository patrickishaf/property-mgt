import type { Property, Unit, Tenant, Payment, DemoAccount } from './types';

export const demoAccounts: DemoAccount[] = [
  { id: 'acc-owner', role: 'owner', name: 'Daniel Okoro' },
  { id: 'acc-mgr-1', role: 'manager', name: 'Grace Mensah', managerId: 'mgr-1' },
  { id: 'acc-tenant-1', role: 'tenant', name: 'Kwame Boateng', tenantId: 't-1' },
  { id: 'acc-tenant-2', role: 'tenant', name: 'Ama Serwaa', tenantId: 't-4' },
];

export const properties: Property[] = [
  {
    id: 'p-1',
    name: 'Adum Shopping Plaza',
    type: 'shopping',
    address: '12 Adum High St, Kumasi',
    managerId: 'mgr-1',
  },
  {
    id: 'p-2',
    name: 'Osu Retail Complex',
    type: 'shopping',
    address: '45 Oxford St, Osu, Accra',
    managerId: 'mgr-1',
  },
  {
    id: 'p-3',
    name: 'East Legon Apartments',
    type: 'apartment',
    address: '8 Boundary Rd, East Legon, Accra',
    managerId: 'mgr-1',
  },
  {
    id: 'p-4',
    name: 'Spintex Garden Flats',
    type: 'apartment',
    address: '120 Spintex Rd, Accra',
    managerId: null,
  },
];

export const units: Unit[] = [
  { id: 'u-1', propertyId: 'p-1', label: 'Shop A1', floor: 0 },
  { id: 'u-2', propertyId: 'p-1', label: 'Shop A2', floor: 0 },
  { id: 'u-3', propertyId: 'p-1', label: 'Shop B1', floor: 1 },
  { id: 'u-4', propertyId: 'p-1', label: 'Shop B2', floor: 1 },
  { id: 'u-5', propertyId: 'p-2', label: 'Unit 101', floor: 0 },
  { id: 'u-6', propertyId: 'p-2', label: 'Unit 102', floor: 0 },
  { id: 'u-7', propertyId: 'p-2', label: 'Unit 201', floor: 1 },
  { id: 'u-8', propertyId: 'p-3', label: 'Apt 1A', floor: 1 },
  { id: 'u-9', propertyId: 'p-3', label: 'Apt 1B', floor: 1 },
  { id: 'u-10', propertyId: 'p-3', label: 'Apt 2A', floor: 2 },
  { id: 'u-11', propertyId: 'p-3', label: 'Apt 2B', floor: 2 },
  { id: 'u-12', propertyId: 'p-4', label: 'Flat 1', floor: 1 },
  { id: 'u-13', propertyId: 'p-4', label: 'Flat 2', floor: 1 },
];

export const tenants: Tenant[] = [
  {
    id: 't-1', name: 'Kwame Boateng', email: 'kwame.b@mail.com', phone: '+233 24 001 0001',
    unitId: 'u-1', propertyId: 'p-1', leaseStart: '2024-01-01', monthlyRent: 3200, dueDayOfMonth: 5,
  },
  {
    id: 't-2', name: 'Adwoa Nyamekye', email: 'adwoa.n@mail.com', phone: '+233 24 001 0002',
    unitId: 'u-2', propertyId: 'p-1', leaseStart: '2023-08-15', monthlyRent: 2800, dueDayOfMonth: 5,
  },
  {
    id: 't-3', name: 'Yaw Asante', email: 'yaw.a@mail.com', phone: '+233 24 001 0003',
    unitId: 'u-3', propertyId: 'p-1', leaseStart: '2024-03-01', monthlyRent: 3500, dueDayOfMonth: 10,
  },
  {
    id: 't-4', name: 'Ama Serwaa', email: 'ama.s@mail.com', phone: '+233 24 001 0004',
    unitId: 'u-8', propertyId: 'p-3', leaseStart: '2023-11-01', monthlyRent: 1800, dueDayOfMonth: 1,
  },
  {
    id: 't-5', name: 'Kofi Mensah', email: 'kofi.m@mail.com', phone: '+233 24 001 0005',
    unitId: 'u-9', propertyId: 'p-3', leaseStart: '2024-02-01', monthlyRent: 1800, dueDayOfMonth: 1,
  },
  {
    id: 't-6', name: 'Esi Owusu', email: 'esi.o@mail.com', phone: '+233 24 001 0006',
    unitId: 'u-10', propertyId: 'p-3', leaseStart: '2024-05-01', monthlyRent: 2200, dueDayOfMonth: 15,
  },
  {
    id: 't-7', name: 'Mensah Otabil', email: 'mensah.o@mail.com', phone: '+233 24 001 0007',
    unitId: 'u-5', propertyId: 'p-2', leaseStart: '2023-09-01', monthlyRent: 3000, dueDayOfMonth: 7,
  },
  {
    id: 't-8', name: 'Akua Fremah', email: 'akua.f@mail.com', phone: '+233 24 001 0008',
    unitId: 'u-6', propertyId: 'p-2', leaseStart: '2024-01-15', monthlyRent: 2600, dueDayOfMonth: 7,
  },
  {
    id: 't-9', name: 'Nana Yaw', email: 'nana.y@mail.com', phone: '+233 24 001 0009',
    unitId: 'u-12', propertyId: 'p-4', leaseStart: '2024-04-01', monthlyRent: 1500, dueDayOfMonth: 3,
  },
  {
    id: 't-10', name: 'Abena Dapaah', email: 'abena.d@mail.com', phone: '+233 24 001 0010',
    unitId: 'u-13', propertyId: 'p-4', leaseStart: '2024-06-01', monthlyRent: 1500, dueDayOfMonth: 3,
  },
];

export const payments: Payment[] = [
  { id: 'pay-1', tenantId: 't-1', amount: 1600, date: '2025-08-03', method: 'bank', reference: 'BNK08821', recordedBy: 'acc-mgr-1' },
  { id: 'pay-2', tenantId: 't-1', amount: 800, date: '2025-08-08', method: 'mobile', reference: 'MOM99221', recordedBy: 'acc-owner' },
  { id: 'pay-3', tenantId: 't-2', amount: 2800, date: '2025-08-04', method: 'bank', reference: 'BNK08822', recordedBy: 'acc-mgr-1' },
  { id: 'pay-4', tenantId: 't-3', amount: 1750, date: '2025-08-09', method: 'cash', reference: 'CSH00451', recordedBy: 'acc-mgr-1' },
  { id: 'pay-5', tenantId: 't-4', amount: 900, date: '2025-08-01', method: 'mobile', reference: 'MOM99230', recordedBy: 'acc-owner' },
  { id: 'pay-6', tenantId: 't-4', amount: 500, date: '2025-08-05', method: 'mobile', reference: 'MOM99231', recordedBy: 'acc-mgr-1' },
  { id: 'pay-7', tenantId: 't-5', amount: 1800, date: '2025-08-01', method: 'bank', reference: 'BNK08830', recordedBy: 'acc-mgr-1' },
  { id: 'pay-8', tenantId: 't-6', amount: 2200, date: '2025-08-14', method: 'card', reference: 'CRD55201', recordedBy: 'acc-owner' },
  { id: 'pay-9', tenantId: 't-7', amount: 1500, date: '2025-08-06', method: 'bank', reference: 'BNK08840', recordedBy: 'acc-mgr-1' },
  { id: 'pay-10', tenantId: 't-8', amount: 2600, date: '2025-08-07', method: 'mobile', reference: 'MOM99240', recordedBy: 'acc-mgr-1' },
  { id: 'pay-11', tenantId: 't-9', amount: 750, date: '2025-08-02', method: 'cash', reference: 'CSH00460', recordedBy: 'acc-owner' },
  { id: 'pay-12', tenantId: 't-10', amount: 1500, date: '2025-08-03', method: 'bank', reference: 'BNK08850', recordedBy: 'acc-mgr-1' },
  { id: 'pay-13', tenantId: 't-1', amount: 3200, date: '2025-07-04', method: 'bank', reference: 'BNK07820', recordedBy: 'acc-owner' },
  { id: 'pay-14', tenantId: 't-4', amount: 1800, date: '2025-07-01', method: 'mobile', reference: 'MOM98230', recordedBy: 'acc-mgr-1' },
  { id: 'pay-15', tenantId: 't-7', amount: 3000, date: '2025-07-06', method: 'bank', reference: 'BNK07840', recordedBy: 'acc-mgr-1' },
];
