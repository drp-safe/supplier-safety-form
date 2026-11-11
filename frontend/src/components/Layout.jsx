import React, { useState } from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import {
  LayoutDashboard, BookOpen, GitPullRequest, FileText, PenSquare,
  Send, BarChart3, ShieldCheck, LogOut, Menu, X,
} from 'lucide-react';
import { useAuth } from '../AuthContext';
import { useLanguage } from '../LanguageContext';

function NavSection({ title, children }) {
  return (
    <div className="mb-4">
      <div className="px-3 pb-1 text-[11px] font-semibold uppercase tracking-wide text-white/40">{title}</div>
      <div className="space-y-1">{children}</div>
    </div>
  );
}

function NavItem({ to, label, icon: Icon, onClick }) {
  return (
    <NavLink
      to={to}
      end={to === '/'}
      onClick={onClick}
      className={({ isActive }) =>
        `flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition ${
          isActive ? 'bg-white/10 text-white border-l-2 border-brand-accent' : 'text-white/70 hover:bg-white/5'
        }`
      }
    >
      <Icon size={18} /> {label}
    </NavLink>
  );
}

export default function Layout() {
  const { user, logout } = useAuth();
  const { t } = useLanguage();
  const [open, setOpen] = useState(false);
  const close = () => setOpen(false);

  return (
    <div className="flex min-h-screen">
      {open && <div className="fixed inset-0 z-30 bg-black/40 lg:hidden" onClick={close} aria-hidden="true" />}

      <aside
        className={`fixed inset-y-0 left-0 z-40 flex w-64 shrink-0 transform flex-col bg-brand-navy text-white transition-transform duration-200 lg:static lg:translate-x-0 ${
          open ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex items-center justify-between px-5 py-5">
          <div className="flex items-center gap-2 text-base font-bold">
            <div className="flex h-7 w-7 items-center justify-center rounded bg-primary text-xs">SF</div>
            {t('common.appName')}
          </div>
          <button onClick={close} className="text-white/70 hover:text-white lg:hidden" aria-label="Fermer le menu">
            <X size={20} />
          </button>
        </div>

        <nav className="flex-1 overflow-y-auto px-3">
          <NavSection title="Pilotage">
            <NavItem to="/" label={t('nav.dashboard')} icon={LayoutDashboard} onClick={close} />
          </NavSection>
          <NavSection title="Questions">
            <NavItem to="/library" label={t('nav.library')} icon={BookOpen} onClick={close} />
            <NavItem to="/proposals" label={t('nav.proposals')} icon={GitPullRequest} onClick={close} />
          </NavSection>
          <NavSection title="Formulaires">
            <NavItem to="/forms" label={t('nav.myForms')} icon={FileText} onClick={close} />
            <NavItem to="/editor" label={t('nav.editor')} icon={PenSquare} onClick={close} />
          </NavSection>
          <NavSection title="Fournisseurs">
            <NavItem to="/solicitations" label={t('nav.solicitations')} icon={Send} onClick={close} />
            <NavItem to="/results" label={t('nav.results')} icon={BarChart3} onClick={close} />
          </NavSection>
          {user?.is_superadmin && (
            <NavSection title="Plateforme">
              <NavItem to="/admin" label={t('nav.administration')} icon={ShieldCheck} onClick={close} />
            </NavSection>
          )}
        </nav>

        <div className="border-t border-white/10 px-3 py-4">
          <div className="truncate px-3 pb-2 text-xs text-white/50">{user?.email}</div>
          <button
            onClick={logout}
            className="flex w-full items-center gap-3 rounded-lg px-3 py-2 text-sm text-white/70 hover:bg-white/5"
          >
            <LogOut size={18} /> {t('nav.logout')}
          </button>
        </div>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-20 flex items-center gap-3 border-b border-neutral-200 bg-white px-4 py-3 lg:hidden">
          <button onClick={() => setOpen(true)} className="text-neutral-600 hover:text-brand-navy" aria-label="Ouvrir le menu">
            <Menu size={22} />
          </button>
          <div className="font-bold text-brand-navy">{t('common.appName')}</div>
        </header>

        <main className="flex-1 overflow-x-hidden">
          <div className="mx-auto max-w-6xl px-4 py-6 lg:px-8 lg:py-8">
            <Outlet />
          </div>
        </main>
      </div>
    </div>
  );
}
