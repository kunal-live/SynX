'use client';

import React from 'react';
import { Peer } from '@/types/synx';
import { RadarScanner } from './RadarScanner';

interface DevicesViewProps {
  peers: Peer[];
  onOpenPairModal: () => void;
  onShowPINModal: () => void;
  onSendToPeer: (peerId: string) => void;
}

function getPlatformIcon(platform?: string) {
  const p = (platform || '').toLowerCase();
  if (p.includes('win')) {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <rect x="3" y="3" width="8" height="8" />
        <rect x="13" y="3" width="8" height="8" />
        <rect x="3" y="13" width="8" height="8" />
        <rect x="13" y="13" width="8" height="8" />
      </svg>
    );
  } else if (p.includes('dar') || p.includes('mac') || p.includes('apple')) {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 2a4 4 0 0 1 4 4c0 3-4 6-4 6s-4-3-4-6a4 4 0 0 1 4-4z" />
        <circle cx="12" cy="14" r="8" />
      </svg>
    );
  } else if (p.includes('linux')) {
    return (
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M12 2a5 5 0 0 0-5 5v5a5 5 0 0 0 10 0V7a5 5 0 0 0-5-5z" />
        <path d="M18 14v4a4 4 0 0 1-8 0v-4" />
      </svg>
    );
  }
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="2" y="3" width="20" height="14" rx="2" />
      <line x1="8" y1="21" x2="16" y2="21" />
      <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
  );
}

export const DevicesView: React.FC<DevicesViewProps> = ({
  peers,
  onOpenPairModal,
  onShowPINModal,
  onSendToPeer
}) => {
  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <div>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '20px', color: 'var(--text-pure)', fontWeight: 800 }}>
            Discovered Mesh Nodes
          </h3>
          <p style={{ fontSize: '13px', color: 'var(--text-med)', marginTop: '2px' }}>
            All active SynX nodes auto-discovered on UDP port 8788.
          </p>
        </div>
        <div style={{ display: 'flex', gap: '12px' }}>
          <button className="btn" onClick={onOpenPairModal}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="8.5" cy="7" r="4" />
              <line x1="20" y1="8" x2="20" y2="14" />
              <line x1="23" y1="11" x2="17" y2="11" />
            </svg>
            Direct IP Connect
          </button>
          <button className="btn btn-primary" onClick={onShowPINModal}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="11" width="18" height="11" rx="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
            Display Pairing PIN
          </button>
        </div>
      </div>

      {/* Radar Topology Card */}
      <div className="hero-mesh-card" style={{ marginBottom: '24px', padding: '24px 28px' }}>
        <div>
          <div className="hero-info-tag">Mesh Topology Status</div>
          <h2 style={{ fontSize: '24px' }}>Local Peer Mesh Network</h2>
          <p style={{ marginBottom: 0 }}>
            Devices on the same network automatically establish bidirectional trust relationships using Ed25519 public keys and fast local sockets.
          </p>
        </div>
        <RadarScanner peers={peers} />
      </div>

      {/* Mesh Node Registry */}
      <div className="panel-card">
        <div className="panel-head">
          <div className="panel-title-group">
            <svg viewBox="0 0 24 24" fill="none" stroke="var(--cyan)" strokeWidth="2">
              <rect x="2" y="3" width="20" height="14" rx="2" />
              <line x1="8" y1="21" x2="16" y2="21" />
            </svg>
            <span className="panel-title">Active Mesh Node Registry ({peers.length})</span>
          </div>
        </div>

        {peers.length === 0 ? (
          <div className="empty-placeholder">
            <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8">
              <rect x="2" y="3" width="20" height="14" rx="2" />
              <line x1="8" y1="21" x2="16" y2="21" />
              <line x1="12" y1="17" x2="12" y2="21" />
            </svg>
            <div>No peers detected on local network.</div>
            <small style={{ color: 'var(--text-dim)', marginTop: '-4px' }}>
              Run SynX on another device connected to the same Wi-Fi or router.
            </small>
          </div>
        ) : (
          peers.map((p) => (
            <div key={p.id} className="device-item-row">
              <div className="device-avatar-box">
                {getPlatformIcon(p.platform)}
              </div>
              <div className="device-details">
                <div className="device-name-line">
                  {p.name || 'SynX Device'}
                  <span className={`badge-tag ${p.trusted ? 'badge-tag-trusted' : 'badge-tag-online'}`}>
                    {p.trusted ? 'Trusted' : 'Discovered'}
                  </span>
                </div>
                <div className="device-address-line">
                  {p.address} • {p.platform || 'LAN Node'}
                </div>
              </div>
              <button className="btn btn-sm btn-primary" onClick={() => onSendToPeer(p.id)}>
                Send
              </button>
            </div>
          ))
        )}
      </div>
    </div>
  );
};
