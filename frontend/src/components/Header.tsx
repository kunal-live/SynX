'use client';

import React from 'react';

interface HeaderProps {
  title: string;
  subtitle: string;
  peerCount: number;
  onOpenPairModal: () => void;
  onBroadcastClick: () => void;
}

export const Header: React.FC<HeaderProps> = ({
  title,
  subtitle,
  peerCount,
  onOpenPairModal,
  onBroadcastClick
}) => {
  return (
    <header className="top-nav">
      <div className="top-titles">
        <h1 id="pageTitle">{title}</h1>
        <p id="pageSubtitle">{subtitle}</p>
      </div>

      <div className="top-actions">
        <div className="mesh-status-indicator">
          <span className="pulse-beacon" style={{ background: 'var(--cyan)', boxShadow: '0 0 8px var(--cyan)' }}></span>
          <span>
            {peerCount > 0 ? `${peerCount} Node${peerCount === 1 ? '' : 's'} Online` : 'Scanning LAN…'}
          </span>
        </div>
        <button className="btn" onClick={onOpenPairModal}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="11" width="18" height="11" rx="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          Pair Device
        </button>
        <button className="btn btn-primary" onClick={onBroadcastClick}>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="22" y1="2" x2="11" y2="13" />
            <polygon points="22 2 15 22 11 13 2 9 22 2" />
          </svg>
          Broadcast Files
        </button>
      </div>
    </header>
  );
};
