'use client';

import React from 'react';

interface SidebarProps {
  currentPage: string;
  onPageChange: (page: string) => void;
  peerCount: number;
  transferCount: number;
  nodeName: string;
  nodeAddress: string;
  onCopyAddress: () => void;
  onOpenFolder: () => void;
}

export const Sidebar: React.FC<SidebarProps> = ({
  currentPage,
  onPageChange,
  peerCount,
  transferCount,
  nodeName,
  nodeAddress,
  onCopyAddress,
  onOpenFolder
}) => {
  return (
    <aside className="sidebar">
      <div className="brand" onClick={() => onPageChange('overview')} style={{ cursor: 'pointer' }}>
        <div className="brand-mark">
          <svg viewBox="0 0 24 24">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5" />
          </svg>
        </div>
        <div className="brand-title">
          <span className="brand-name">SynX</span>
          <span className="brand-badge">Next.js Core</span>
        </div>
      </div>

      <nav className="nav-menu">
        <button
          className={`nav-item ${currentPage === 'overview' ? 'active' : ''}`}
          onClick={() => onPageChange('overview')}
        >
          <svg viewBox="0 0 24 24" fill="none" strokeWidth="2">
            <rect x="3" y="3" width="7" height="7" />
            <rect x="14" y="3" width="7" height="7" />
            <rect x="14" y="14" width="7" height="7" />
            <rect x="3" y="14" width="7" height="7" />
          </svg>
          <span>Mesh Overview</span>
        </button>

        <button
          className={`nav-item ${currentPage === 'devices' ? 'active' : ''}`}
          onClick={() => onPageChange('devices')}
        >
          <svg viewBox="0 0 24 24" fill="none" strokeWidth="2">
            <rect x="2" y="3" width="20" height="14" rx="2" />
            <line x1="8" y1="21" x2="16" y2="21" />
            <line x1="12" y1="17" x2="12" y2="21" />
          </svg>
          <span>Connected Nodes</span>
          <span className="nav-badge">{peerCount}</span>
        </button>

        <button
          className={`nav-item ${currentPage === 'transfers' ? 'active' : ''}`}
          onClick={() => onPageChange('transfers')}
        >
          <svg viewBox="0 0 24 24" fill="none" strokeWidth="2">
            <polyline points="17 1 21 5 17 9" />
            <path d="M3 11V9a4 4 0 0 1 4-4h14" />
            <polyline points="7 23 3 19 7 15" />
            <path d="M21 13v2a4 4 0 0 1-4 4H3" />
          </svg>
          <span>Transfer Streams</span>
          <span className="nav-badge">{transferCount}</span>
        </button>

        <button
          className={`nav-item ${currentPage === 'explorer' ? 'active' : ''}`}
          onClick={() => onPageChange('explorer')}
        >
          <svg viewBox="0 0 24 24" fill="none" strokeWidth="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
          <span>Shared Storage</span>
        </button>

        <button
          className={`nav-item ${currentPage === 'settings' ? 'active' : ''}`}
          onClick={() => onPageChange('settings')}
        >
          <svg viewBox="0 0 24 24" fill="none" strokeWidth="2">
            <circle cx="12" cy="12" r="3" />
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
          </svg>
          <span>Node Settings</span>
        </button>
      </nav>

      <div className="sidebar-spacer"></div>

      {/* Local Node Widget */}
      <div className="node-widget">
        <div className="node-widget-head">
          <span className="node-status-pill">
            <span className="pulse-beacon"></span> Node Online
          </span>
          <button className="btn-icon-copy" onClick={onOpenFolder} title="Open shared folder in OS explorer">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
              <polyline points="15 3 21 3 21 9" />
              <line x1="10" y1="14" x2="21" y2="3" />
            </svg>
          </button>
        </div>
        <div className="node-title">{nodeName || 'SynX Node'}</div>
        <div className="node-ip-row">
          <span>{nodeAddress || '0.0.0.0:8787'}</span>
          <button className="btn-icon-copy" onClick={onCopyAddress} title="Copy Address">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="9" y="9" width="13" height="13" rx="2" />
              <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
            </svg>
          </button>
        </div>
      </div>
    </aside>
  );
};
