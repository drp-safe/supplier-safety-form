import React, { useEffect, useState } from 'react';
import { useLanguage } from '../LanguageContext';
import { adminAPI } from '../api';
import { Card, Field, Input, Select, Button, Alert } from '../components/ui';

const TABS = ['email', 'sectors', 'oauth', 'db', 'migrations'];

export default function Admin() {
  const { t } = useLanguage();
  const [tab, setTab] = useState('email');

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-brand-navy">{t('admin.title')}</h1>

      <div className="mb-6 flex gap-1 border-b border-neutral-200">
        {TABS.map((key) => (
          <button
            key={key}
            onClick={() => setTab(key)}
            className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px transition ${
              tab === key ? 'border-primary text-primary' : 'border-transparent text-neutral-500 hover:text-neutral-700'
            }`}
          >
            {t(`admin.tab${key[0].toUpperCase()}${key.slice(1)}`)}
          </button>
        ))}
      </div>

      {tab === 'email' && <EmailConfigTab />}
      {tab === 'sectors' && <SectorsTab />}
      {tab === 'oauth' && <OAuthTab />}
      {tab === 'db' && <DBInfoTab />}
      {tab === 'migrations' && <MigrationsTab />}
    </div>
  );
}

function EmailConfigTab() {
  const { t } = useLanguage();
  const [form, setForm] = useState(null);
  const [testEmail, setTestEmail] = useState('');
  const [message, setMessage] = useState(null);

  useEffect(() => {
    adminAPI.getEmailConfig().then((res) => setForm(res.data)).catch(() => setForm({
      smtp_host: '', smtp_port: 587, smtp_username: '', smtp_password: '',
      smtp_encryption: 'tls', from_address: '', from_name: '',
    }));
  }, []);

  const set = (field) => (e) => setForm((f) => ({ ...f, [field]: e.target.value }));

  const save = async () => {
    setMessage(null);
    try {
      await adminAPI.updateEmailConfig(form);
      setMessage({ type: 'success', text: t('common.save') });
    } catch (err) {
      setMessage({ type: 'danger', text: err.response?.data?.message });
    }
  };

  const sendTest = async () => {
    setMessage(null);
    try {
      await adminAPI.testEmailConfig(testEmail);
      setMessage({ type: 'success', text: t('admin.sendTest') });
    } catch (err) {
      setMessage({ type: 'danger', text: err.response?.data?.message });
    }
  };

  if (!form) return null;

  return (
    <Card className="max-w-xl space-y-4 p-6">
      {message && <Alert type={message.type}>{message.text}</Alert>}
      <div className="grid grid-cols-2 gap-3">
        <Field label={t('admin.emailHost')}>
          <Input value={form.smtp_host} onChange={set('smtp_host')} />
        </Field>
        <Field label={t('admin.emailPort')}>
          <Input type="number" value={form.smtp_port} onChange={(e) => setForm((f) => ({ ...f, smtp_port: Number(e.target.value) }))} />
        </Field>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Field label={t('admin.emailUsername')}>
          <Input value={form.smtp_username} onChange={set('smtp_username')} />
        </Field>
        <Field label={t('admin.emailPassword')}>
          <Input type="password" value={form.smtp_password} onChange={set('smtp_password')} />
        </Field>
      </div>
      <Field label={t('admin.emailEncryption')}>
        <Select value={form.smtp_encryption} onChange={set('smtp_encryption')}>
          <option value="tls">STARTTLS</option>
          <option value="ssl">SSL</option>
          <option value="none">{t('admin.emailEncryptionNone')}</option>
        </Select>
      </Field>
      <div className="grid grid-cols-2 gap-3">
        <Field label={t('admin.emailFromAddress')}>
          <Input type="email" value={form.from_address} onChange={set('from_address')} />
        </Field>
        <Field label={t('admin.emailFromName')}>
          <Input value={form.from_name} onChange={set('from_name')} />
        </Field>
      </div>
      <Button onClick={save}>{t('common.save')}</Button>

      <div className="flex items-end gap-2 border-t border-neutral-100 pt-4">
        <Field label={t('admin.testEmailLabel')} className="flex-1">
          <Input type="email" value={testEmail} onChange={(e) => setTestEmail(e.target.value)} />
        </Field>
        <Button variant="secondary" onClick={sendTest} disabled={!testEmail}>{t('admin.sendTest')}</Button>
      </div>
    </Card>
  );
}

function SectorsTab() {
  const { t } = useLanguage();
  const [sectors, setSectors] = useState([]);
  const [draft, setDraft] = useState({ code: '', label: '', display_order: 0 });

  const load = () => adminAPI.listSectors().then((res) => setSectors(res.data || []));
  useEffect(() => { load(); }, []);

  const add = async () => {
    if (!draft.code || !draft.label) return;
    await adminAPI.createSector({ ...draft, active: true });
    setDraft({ code: '', label: '', display_order: 0 });
    load();
  };

  const toggleActive = async (sector) => {
    await adminAPI.updateSector(sector.code, { ...sector, active: !sector.active });
    load();
  };

  return (
    <Card className="max-w-2xl p-6">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-neutral-500">
            <th className="pb-2">{t('admin.sectorCode')}</th>
            <th className="pb-2">{t('admin.sectorLabel')}</th>
            <th className="pb-2">{t('admin.sectorActive')}</th>
          </tr>
        </thead>
        <tbody>
          {sectors.map((s) => (
            <tr key={s.code} className="border-t border-neutral-100">
              <td className="py-2 font-mono text-xs text-neutral-500">{s.code}</td>
              <td className="py-2">{s.label}</td>
              <td className="py-2">
                <input type="checkbox" checked={s.active} onChange={() => toggleActive(s)} />
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      <div className="mt-4 flex items-end gap-2 border-t border-neutral-100 pt-4">
        <Field label={t('admin.sectorCode')}>
          <Input value={draft.code} onChange={(e) => setDraft((d) => ({ ...d, code: e.target.value }))} />
        </Field>
        <Field label={t('admin.sectorLabel')}>
          <Input value={draft.label} onChange={(e) => setDraft((d) => ({ ...d, label: e.target.value }))} />
        </Field>
        <Button onClick={add}>{t('admin.addSector')}</Button>
      </div>
    </Card>
  );
}

function OAuthTab() {
  const { t } = useLanguage();
  const [config, setConfig] = useState(null);
  useEffect(() => { adminAPI.getOAuthConfig().then((res) => setConfig(res.data)); }, []);
  if (!config) return null;
  return (
    <Card className="max-w-xl space-y-3 p-6">
      <Alert type="info">{t('admin.oauthHint')}</Alert>
      <Field label={t('admin.oauthIssuer')}><Input value={config.issuer_url} readOnly /></Field>
      <Field label={t('admin.oauthClient')}><Input value={config.client_id} readOnly /></Field>
    </Card>
  );
}

function DBInfoTab() {
  const { t } = useLanguage();
  const [info, setInfo] = useState(null);
  useEffect(() => { adminAPI.getDBInfo().then((res) => setInfo(res.data)); }, []);
  if (!info) return null;
  return (
    <Card className="max-w-xl space-y-3 p-6">
      <Alert type="info">{t('admin.dbHint')}</Alert>
      <div className="grid grid-cols-2 gap-3">
        <Field label={t('admin.dbHost')}><Input value={info.host} readOnly /></Field>
        <Field label={t('admin.dbPort')}><Input value={info.port} readOnly /></Field>
      </div>
      <div className="grid grid-cols-2 gap-3">
        <Field label={t('admin.dbName')}><Input value={info.name} readOnly /></Field>
        <Field label={t('admin.dbDriver')}><Input value={info.driver} readOnly /></Field>
      </div>
    </Card>
  );
}

function MigrationsTab() {
  const { t } = useLanguage();
  const [migrations, setMigrations] = useState([]);
  useEffect(() => { adminAPI.listMigrations().then((res) => setMigrations(res.data || [])); }, []);
  return (
    <Card className="p-6">
      <table className="w-full text-sm">
        <thead>
          <tr className="text-left text-neutral-500">
            <th className="pb-2">{t('admin.migrationsVersion')}</th>
            <th className="pb-2">{t('admin.migrationsDescription')}</th>
            <th className="pb-2">{t('admin.migrationsAppliedAt')}</th>
            <th className="pb-2">{t('admin.migrationsStatus')}</th>
          </tr>
        </thead>
        <tbody>
          {migrations.map((m) => (
            <tr key={m.version} className="border-t border-neutral-100">
              <td className="py-2 font-mono text-xs">{m.version}</td>
              <td className="py-2">{m.description}</td>
              <td className="py-2 text-neutral-500">{m.applied_at}</td>
              <td className="py-2">{m.success ? '✅' : '❌'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
