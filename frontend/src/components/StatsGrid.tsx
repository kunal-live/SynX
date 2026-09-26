'use client';

import React from 'react';
import { Metrics } from '@/types/synx';

interface StatsGridProps {
  peerCount: number;
  transferCount: number;
  metrics: Metrics;
}

function formatBytes(bytes: number | undefined): string {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export const StatsGrid: React.FC<StatsGridProps> = ({
  peerCount,
  transferCount,
  metrics
}) => {
  return (
    <div className="stats-grid">
      <div className="stat-box">
        <div className="stat-box-header">
          <span className="stat-label">Mesh Nodes</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
            <path d="M2 12h20" />
          </svg>
        </div>
        <div className="stat-val" style={{ color: 'var(--emerald)' }}>{peerCount}</div>
        <div className="stat-sub">Active on local network</div>
      </div>

      <div className="stat-box">
        <div className="stat-box-header">
          <span className="stat-label">Active Streams</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <polyline points="17 1 21 5 17 9" />
            <path d="M3 11V9a4 4 0 0 1 4-4h14" />
          </svg>
        </div>
        <div className="stat-val" style={{ color: 'var(--cyan)' }}>{transferCount}</div>
        <div className="stat-sub">Chunk transfer queues</div>
      </div>

      <div className="stat-box">
        <div className="stat-box-header">
          <span className="stat-label">Outbound Sent</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="12" y1="19" x2="12" y2="5" />
            <polyline points="5 12 12 5 19 12" />
          </svg>
        </div>
        <div className="stat-val">{formatBytes(metrics.bytes_sent)}</div>
        <div className="stat-sub">Direct peer broadcasts</div>
      </div>

      <div className="stat-box">
        <div className="stat-box-header">
          <span className="stat-label">Inbound Verified</span>
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="12" y1="5" x2="12" y2="19" />
            <polyline points="19 12 12 19 5 12" />
          </svg>
        </div>
        <div className="stat-val">{formatBytes(metrics.bytes_received)}</div>
        <div className="stat-sub">Checksum validated</div>
      </div>
    </div>
  );
};
