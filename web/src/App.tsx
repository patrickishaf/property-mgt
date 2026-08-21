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
import { CreateAccount } from './components/CreateAccount';
import { Building, Loader2 } from 'lucide-react';

function Dashboard() {
  const { account, authReady } = useApp();
  const [page, setPage] = useState<Page>('overview');
  const [selectedProperty, setSelectedProperty] = useState<string | null>(null);
  const [selectedTenant, setSelectedTenant] = useState<string | null>(null);

  if (!authReady) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-gray-50 dark:bg-gray-950">
        <div className="flex flex-col items-center gap-3">
          <div className="w-10 h-10 bg-brand-600 flex items-center justify-center">
            <Building size={20} className="text-white" />
          </div>
          <Loader2 size={20} className="animate-spin text-gray-400" />
        </div>
      </div>
    );
  }

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
      case 'create-account':
        return <CreateAccount />;
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
