import React from 'react';
import { useLanguage } from '../LanguageContext';
import { Card, EmptyState } from '../components/ui';

export default function ComingSoon({ title }) {
  const { t } = useLanguage();
  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-brand-navy">{title}</h1>
      <Card>
        <EmptyState>
          <p className="text-base font-medium text-neutral-600">{t('common.comingSoonTitle')}</p>
          <p className="mt-1 text-sm">{t('common.comingSoonBody')}</p>
        </EmptyState>
      </Card>
    </div>
  );
}
