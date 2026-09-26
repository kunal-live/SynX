'use client';

import React from 'react';

interface SettingsViewProps {
  sharedDir: string;
  nodeName: string;
  token: string;
  address: string;
  onBrowseFolder: () => void;
  onOpenFolder: () => void;
  onCopyToken: () => void;
  onSavePreferences: () => void;
}

export const SettingsView: React.FC<SettingsViewProps> = ({
  sharedDir,
  nodeName,
  token,
  address,
  onBrowseFolder,
  onOpenFolder,
  onCopyToken,
  onSavePreferences
}) => {
  return (
    <div className="settings-wrap">
      {/* Storage & Paths */}
      <div className="settings-card">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="var(--cyan)" strokeWidth="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
          </svg>
          Shared Storage Directory
        </h3>
        <div className="field-block">
          <label className="field-label">Default Transfer Destination</label>
          <div className="field-input-row">
            <input className="input-control" value={sharedDir || '~/SynX'} readOnly />
            <button className="btn" onClick={onBrowseFolder}>Browse…</button>
            <button className="btn" onClick={onOpenFolder}>Open</button>
          </div>
        </div>
      </div>

      {/* Node Identity & Credentials */}
      <div className="settings-card">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="var(--violet)" strokeWidth="2">
            <rect x="3" y="11" width="18" height="11" rx="2" />
            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
          </svg>
          Node Identity &amp; Security Credentials
        </h3>
        <div className="field-block">
          <label className="field-label">Node Hostname</label>
          <input className="input-control" value={nodeName || 'SynX'} readOnly />
        </div>
        <div className="field-block">
          <label className="field-label">Pairing Security Token</label>
          <div className="field-input-row">
            <input className="input-control" value={token || ''} readOnly />
            <button className="btn" onClick={onCopyToken}>Copy Token</button>
          </div>
        </div>
        <div className="field-block">
          <label className="field-label">LAN Listening Socket</label>
          <input className="input-control" value={address || '0.0.0.0:8787'} readOnly />
        </div>
      </div>

      {/* Bandwidth & Network Tuning */}
      <div className="settings-card">
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="var(--emerald)" strokeWidth="2">
            <line x1="12" y1="19" x2="12" y2="5" />
            <polyline points="5 12 12 5 19 12" />
          </svg>
          Bandwidth &amp; Network Tuning
        </h3>
        <div className="field-block">
          <label className="field-label">Max Transfer Throughput Limit</label>
          <select className="input-control" defaultValue="unlimited">
            <option value="unlimited">Unlimited (Full LAN Wire Speed)</option>
            <option value="100">100 MB/s (High Bandwidth)</option>
            <option value="50">50 MB/s (Standard LAN)</option>
            <option value="20">20 MB/s (Low Congestion Mode)</option>
          </select>
        </div>
        <div style={{ marginTop: '18px' }}>
          <button className="btn btn-primary" onClick={onSavePreferences}>
            Save Configuration
          </button>
        </div>
      </div>
    </div>
  );
};
