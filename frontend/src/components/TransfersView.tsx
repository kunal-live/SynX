'use client';

import React, { useEffect, useState } from 'react';
import { Transfer, TransferHistory } from '@/types/synx';
import { synxClient } from '@/lib/synx-client';

interface TransfersViewProps {
  transfers: Transfer[];
  onCancelTransfer?: (id: string) => void;
}

function formatBytes(bytes: number | undefined): string {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

export const TransfersView: React.FC<TransfersViewProps> = ({ transfers, onCancelTransfer }) => {
  const [history, setHistory] = useState<TransferHistory[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);

  const loadHistory = async () => {
    setLoadingHistory(true);
    try {
      const data = await synxClient.getHistory(50);
      setHistory(data || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoadingHistory(false);
    }
  };

  useEffect(() => {
    loadHistory();
  }, []);

  return (
    <div>
      {/* Active Streams Card */}
      <div className="panel-card" style={{ marginBottom: '24px' }}>
        <div className="panel-head">
          <div className="panel-title-group">
            <svg viewBox="0 0 24 24" fill="none" stroke="var(--cyan)" strokeWidth="2">
              <polyline points="17 1 21 5 17 9" />
              <path d="M3 11V9a4 4 0 0 1 4-4h14" />
            </svg>
            <span className="panel-title">
              Active Streaming Chunk Queues ({transfers.length})
            </span>
          </div>
        </div>

        {transfers.length === 0 ? (
          <div className="empty-placeholder">
            <svg viewBox="0 0 24 24" fill="none" strokeWidth="1.8">
              <circle cx="12" cy="12" r="10" />
              <line x1="12" y1="8" x2="12" y2="12" />
              <line x1="12" y1="16" x2="12.01" y2="16" />
            </svg>
            <div>No active transfers in progress. Drop files to begin high-speed LAN streaming.</div>
          </div>
        ) : (
          transfers.map((t) => (
            <div key={t.id} className="transfer-card">
              <div className="transfer-top">
                <div className="transfer-name" title={t.name}>{t.name}</div>
                <div className="transfer-percentage">{Math.round(t.progress || 0)}%</div>
              </div>
              <div className="progress-container">
                <div
                  className="progress-fill"
                  style={{ width: `${Math.min(100, Math.max(0, t.progress || 0))}%` }}
                ></div>
              </div>
              <div className="transfer-meta-row">
                <span>{t.status || 'streaming'} • {formatBytes(t.done)} / {formatBytes(t.size)}</span>
                <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                  <span>{t.speed || 'LAN Stream'} {t.eta ? `• ${t.eta}s rem` : ''}</span>
                  {onCancelTransfer && (
                    <button
                      className="btn btn-sm"
                      style={{ padding: '3px 8px', fontSize: '11px', color: 'var(--coral)' }}
                      onClick={() => onCancelTransfer(t.id)}
                    >
                      Cancel
                    </button>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>

      {/* SQLite Transfer History Card */}
      <div className="panel-card">
        <div className="panel-head">
          <div className="panel-title-group">
            <svg viewBox="0 0 24 24" fill="none" stroke="var(--emerald)" strokeWidth="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
            </svg>
            <span className="panel-title">Transfer History (SQLite Committed)</span>
          </div>
          <button className="btn btn-sm" onClick={loadHistory} disabled={loadingHistory}>
            {loadingHistory ? 'Loading…' : 'Refresh Logs'}
          </button>
        </div>

        {history.length === 0 ? (
          <div className="empty-placeholder">
            No historical transfers committed yet. Completed transfers are stored in local SQLite database.
          </div>
        ) : (
          <table className="data-table">
            <thead>
              <tr>
                <th>File Name</th>
                <th>Size</th>
                <th>Direction</th>
                <th>Integrity Status</th>
                <th>Timestamp</th>
              </tr>
            </thead>
            <tbody>
              {history.map((h) => (
                <tr key={h.id}>
                  <td><strong>{h.file_name}</strong></td>
                  <td>{formatBytes(h.size)}</td>
                  <td>
                    <span className={`badge-tag ${h.direction === 'send' ? 'badge-tag-trusted' : 'badge-tag-online'}`}>
                      {h.direction.toUpperCase()}
                    </span>
                  </td>
                  <td>
                    <span style={{ color: 'var(--emerald)', fontWeight: 700 }}>
                      {h.status}
                    </span>
                  </td>
                  <td style={{ color: 'var(--text-med)', fontFamily: 'var(--font-mono)', fontSize: '12px' }}>
                    {new Date(h.completed_at * 1000).toLocaleString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};
