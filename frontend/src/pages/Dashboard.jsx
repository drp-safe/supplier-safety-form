import React from 'react';
import { Send, CheckCircle2, Clock, GitPullRequest } from 'lucide-react';
import { useLanguage } from '../LanguageContext';
import { useAuth } from '../AuthContext';
import { Card } from '../components/ui';

function StatCard({ icon: Icon, label, value }) {
  return (
    <Card className="p-5">
      <div className="mb-2 flex items-center gap-2 text-neutral-400">
        <Icon size={16} />
        <span className="text-sm">{label}</span>
      </div>
      <div className="text-2xl font-bold text-brand-navy">{value}</div>
    </Card>
  );
}

export default function Dashboard() {
  const { t } = useLanguage();
  const { user } = useAuth();

  return (
    <div>
      <h1 className="mb-1 text-2xl font-bold text-brand-navy">{t('dashboard.title')}</h1>
      <p className="mb-6 text-neutral-500">{t('dashboard.welcome', { name: user?.first_name || user?.name || '' })}</p>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
        <StatCard icon={Send} label={t('dashboard.suppliersSolicited')} value={0} />
        <StatCard icon={CheckCircle2} label={t('dashboard.responsesReceived')} value={0} />
        <StatCard icon={Clock} label={t('dashboard.late')} value={0} />
        <StatCard icon={GitPullRequest} label={t('dashboard.correctionsToReview')} value={0} />
      </div>
    </div>
  );
}
