import React from 'react';

const variants = {
  primary: 'bg-primary text-white hover:bg-primary-dark',
  secondary: 'bg-white text-neutral-700 border border-neutral-300 hover:bg-neutral-50',
  danger: 'bg-danger text-white hover:bg-danger-dark',
  ghost: 'text-neutral-600 hover:bg-neutral-100',
};

export function Button({ variant = 'primary', className = '', disabled, children, ...props }) {
  return (
    <button
      disabled={disabled}
      className={`inline-flex items-center justify-center gap-2 rounded-lg px-4 py-2 text-sm font-medium
        transition disabled:opacity-50 disabled:cursor-not-allowed ${variants[variant]} ${className}`}
      {...props}
    >
      {children}
    </button>
  );
}

export function Card({ className = '', children }) {
  return <div className={`card-base ${className}`}>{children}</div>;
}

export function Field({ label, children, hint, className = '' }) {
  return (
    <label className={`block ${className}`}>
      {label && <span className="mb-1 block text-sm font-medium text-neutral-700">{label}</span>}
      {children}
      {hint && <span className="mt-1 block text-xs text-neutral-400">{hint}</span>}
    </label>
  );
}

export function Input({ className = '', ...props }) {
  return <input className={`input-base ${className}`} {...props} />;
}

export function Select({ className = '', children, ...props }) {
  return (
    <select className={`input-base ${className}`} {...props}>
      {children}
    </select>
  );
}

export function Badge({ children, className = '' }) {
  return (
    <span className={`inline-block rounded-full px-2.5 py-0.5 text-xs font-medium bg-neutral-100 text-neutral-600 ${className}`}>
      {children}
    </span>
  );
}

export function Alert({ type = 'info', children }) {
  const map = {
    info: 'bg-secondary-light text-secondary-dark',
    success: 'bg-success-light text-success-dark',
    danger: 'bg-danger-light text-danger-dark',
  };
  return <div className={`rounded-lg px-4 py-3 text-sm ${map[type]}`}>{children}</div>;
}

export function Spinner() {
  return <div className="py-10 text-center text-neutral-400">Chargement…</div>;
}

export function EmptyState({ children }) {
  return <div className="py-12 text-center text-neutral-400">{children}</div>;
}
