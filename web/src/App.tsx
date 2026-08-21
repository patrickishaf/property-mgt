import { useState } from 'react';
import { AppProvider, useApp } from './context';
import { Login } from './components/Login';
import { Layout, type Page } from './components/Layout';
import { Overview } from './components/Overview';
import { PropertiesList, PropertyDetail } from './components/Properties';
import { TenantsList, TenantDetail } from './components/Tenants';
import { PaymentsLog } from './components/PaymentsLog';
import { DueDates } from './components/DueDates';
import { TenantHome } from './components/TenantHome';

function Dashboard() {
  const { account } = useApp();
  const [page, setPage] = useState<Page>('overview');
  const [selectedProperty, setSelectedProperty] = useState<string | null>(null);
  const [selectedTenant, setSelectedTenant] = useState<string | null>(null);

  if (!account) return <Login />;

  const isTenant = account.role === 'tenant';

  const navigate = (p: Page) => {
    setSelectedProperty(null);
    setSelectedTenant(null);
    setPage(p);
  };

  const renderPage = () => {
    if (isTenant) {
      return <TenantHome />;
    }

    // Tenant detail view takes priority
    if (selectedTenant) {
      return <TenantDetail tenantId={selectedTenant} onBack={() => setSelectedTenant(null)} />;
    }

    // Property detail view
    if (page === 'properties' && selectedProperty) {
      return (
        <PropertyDetail
          propertyId={selectedProperty}
          onBack={() => setSelectedProperty(null)}
          onSelectTenant={(id) => setSelectedTenant(id)}
        />
      );
    }

    switch (page) {
      case 'overview':
        return <Overview onNavigate={navigate} />;
      case 'properties':
        return <PropertiesList onSelectProperty={(id) => setSelectedProperty(id)} />;
      case 'tenants':
        return <TenantsList onSelectTenant={(id) => setSelectedTenant(id)} />;
      case 'payments':
        return <PaymentsLog />;
      case 'due-dates':
        return <DueDates />;
      default:
        return <Overview onNavigate={navigate} />;
    }
  };

  const currentPage: Page = isTenant ? 'tenant-home' : page;

  return (
    <Layout current={currentPage} onNavigate={navigate}>
      {renderPage()}
    </Layout>
  );
}

function App() {
  return (
    <AppProvider>
      <Dashboard />
    </AppProvider>
  );
}

export default App;
