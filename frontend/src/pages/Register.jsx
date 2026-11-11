import React, { useEffect, useState } from 'react';
import { useLanguage } from '../LanguageContext';
import { useAuth } from '../AuthContext';
import { registrationAPI, referenceAPI } from '../api';
import { Card, Field, Input, Select, Button, Alert } from '../components/ui';

export default function Register() {
  const { t, language } = useLanguage();
  const { login } = useAuth();
  const [sectors, setSectors] = useState([]);
  const [form, setForm] = useState({ first_name: '', last_name: '', email: '', org_name: '', sector_code: '' });
  const [submitting, setSubmitting] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    referenceAPI.listSectors().then((res) => setSectors(res.data || [])).catch(() => setSectors([]));
  }, []);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      await registrationAPI.register({ ...form, language });
      setSuccess(true);
    } catch (err) {
      setError(err.response?.data?.message || t('register.errorGeneric'));
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center bg-neutral-50 px-4 py-10">
      <Card className="w-full max-w-md p-8">
        <div className="mb-6 flex items-center gap-2 text-lg font-bold text-brand-navy">
          <div className="flex h-8 w-8 items-center justify-center rounded bg-primary text-sm text-white">SF</div>
          {t('common.appName')}
        </div>

        {success ? (
          <Alert type="success">{t('register.success')}</Alert>
        ) : (
          <form onSubmit={handleSubmit} className="space-y-4">
            <h1 className="text-xl font-bold text-brand-navy">{t('register.title')}</h1>
            <p className="text-sm text-neutral-500">{t('register.subtitle')}</p>

            {error && <Alert type="danger">{error}</Alert>}

            <div className="grid grid-cols-2 gap-3">
              <Field label={t('register.firstName')}>
                <Input value={form.first_name} onChange={set('first_name')} required />
              </Field>
              <Field label={t('register.lastName')}>
                <Input value={form.last_name} onChange={set('last_name')} required />
              </Field>
            </div>

            <Field label={t('register.email')}>
              <Input type="email" value={form.email} onChange={set('email')} required />
            </Field>

            <Field label={t('register.orgName')}>
              <Input value={form.org_name} onChange={set('org_name')} required />
            </Field>

            <Field label={t('register.sector')}>
              <Select value={form.sector_code} onChange={set('sector_code')} required>
                <option value="">{t('register.sectorPlaceholder')}</option>
                {sectors.map((s) => (
                  <option key={s.code} value={s.code}>{s.label}</option>
                ))}
              </Select>
            </Field>

            <Button type="submit" className="w-full" disabled={submitting}>
              {t('register.submit')}
            </Button>

            <p className="text-center text-sm text-neutral-500">
              {t('register.alreadyAccount')}{' '}
              <button type="button" onClick={() => login()} className="font-medium text-primary hover:underline">
                {t('register.login')}
              </button>
            </p>
          </form>
        )}
      </Card>
    </div>
  );
}
