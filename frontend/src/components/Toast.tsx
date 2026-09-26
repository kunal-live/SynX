'use client';

import React from 'react';

interface ToastProps {
  message: string | null;
}

export const Toast: React.FC<ToastProps> = ({ message }) => {
  if (!message) return null;

  return (
    <div className="floating-toast" style={{ display: 'flex' }}>
      <span className="pulse-beacon" style={{ background: 'var(--cyan)', boxShadow: '0 0 10px var(--cyan)' }}></span>
      <span>{message}</span>
    </div>
  );
};
