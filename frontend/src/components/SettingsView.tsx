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

      {/* Privacy & Data Transparency */}
      <div className="settings-card" style={{ borderColor: 'rgba(0, 240, 255, 0.2)' }}>
        <h3>
          <svg viewBox="0 0 24 24" width="20" height="20" fill="none" stroke="var(--cyan)" strokeWidth="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
          </svg>
          Privacy &amp; Data Transparency
        </h3>
        <p style={{ fontSize: '12px', color: 'var(--text-med)', lineHeight: 1.5, marginBottom: '14px' }}>
          SynX is 100% decentralized and local-first. We collect <strong>zero telemetry</strong>, operate <strong>zero cloud tracking servers</strong>, and use <strong>no third-party embeds</strong>. All transfers stream directly peer-to-peer over your local area network.
        </p>
        <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '10px', marginBottom: '12px' }}>
          <div style={{ background: 'rgba(255, 255, 255, 0.03)', border: '1px solid var(--border-dim)', borderRadius: '8px', padding: '10px' }}>
            <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--cyan)', marginBottom: '4px' }}>🛡️ Zero Telemetry</div>
            <div style={{ fontSize: '11px', color: 'var(--text-dim)', lineHeight: 1.4 }}>No Google Analytics, Sentry, or user tracking SDKs.</div>
          </div>
          <div style={{ background: 'rgba(255, 255, 255, 0.03)', border: '1px solid var(--border-dim)', borderRadius: '8px', padding: '10px' }}>
            <div style={{ fontSize: '11px', fontWeight: 600, color: 'var(--violet)', marginBottom: '4px' }}>🍪 Zero Tracking Cookies</div>
            <div style={{ fontSize: '11px', color: 'var(--text-dim)', lineHeight: 1.4 }}>Client-side localStorage is used only for UI theme preferences.</div>
          </div>
        </div>
        <div style={{ fontSize: '11px', color: 'var(--text-dim)', borderTop: '1px solid var(--border-dim)', paddingTop: '10px' }}>
          Open-Source Compliance: <span style={{ color: 'var(--cyan)' }}>PRIVACY.md</span> • <span style={{ color: 'var(--cyan)' }}>TERMS.md</span> • <span style={{ color: 'var(--cyan)' }}>COOKIE_POLICY.md</span>
        </div>
      </div>
    </div>
  );
};
