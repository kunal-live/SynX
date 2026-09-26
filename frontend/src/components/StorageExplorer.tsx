'use client';

import React, { useEffect, useState } from 'react';
import { SharedFile } from '@/types/synx';
import { synxClient } from '@/lib/synx-client';

interface StorageExplorerProps {
  sharedDir: string;
  onOpenFolder: () => void;
  onSendFile: (path: string) => void;
  onNotify: (msg: string) => void;
}

function formatBytes(bytes: number | undefined): string {
  if (!bytes || bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
}

function getFileTypeIcon(name: string) {
  const ext = (name.split('.').pop() || '').toLowerCase();
  if (['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg'].includes(ext)) {
    return (
      <div className="file-type-icon icon-image">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <circle cx="8.5" cy="8.5" r="1.5" />
          <polyline points="21 15 16 10 5 21" />
        </svg>
      </div>
    );
  }
  if (['zip', 'tar', 'gz', '7z', 'rar', 'exe', 'iso'].includes(ext)) {
    return (
      <div className="file-type-icon icon-archive">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="3" y="3" width="18" height="18" rx="2" />
          <line x1="12" y1="3" x2="12" y2="21" />
        </svg>
      </div>
    );
  }
  if (['go', 'js', 'ts', 'html', 'css', 'json', 'py', 'rs', 'cpp', 'c', 'java', 'md'].includes(ext)) {
    return (
      <div className="file-type-icon icon-code">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <polyline points="16 18 22 12 16 6" />
          <polyline points="8 6 2 12 8 18" />
        </svg>
      </div>
    );
  }
  if (['pdf', 'docx', 'doc', 'txt', 'rtf'].includes(ext)) {
    return (
      <div className="file-type-icon icon-doc">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
        </svg>
      </div>
    );
  }
  return (
    <div className="file-type-icon icon-default">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <polyline points="14 2 14 8 20 8" />
      </svg>
    </div>
  );
}

export const StorageExplorer: React.FC<StorageExplorerProps> = ({
  sharedDir,
  onOpenFolder,
  onSendFile,
  onNotify
}) => {
  const [files, setFiles] = useState<SharedFile[]>([]);
  const [loading, setLoading] = useState(false);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const data = await synxClient.getSharedFiles('');
      setFiles(data || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadFiles();
  }, []);

  const handleDelete = async (path: string) => {
    if (!confirm(`Are you sure you want to delete "${path}"?`)) return;
    try {
      await synxClient.deleteSharedFile(path);
      onNotify('File removed from shared storage');
      loadFiles();
    } catch (e: any) {
      onNotify('Failed to delete file: ' + e.message);
    }
  };

  return (
    <div className="panel-card">
      <div className="panel-head">
        <div>
          <div className="panel-title-group">
            <svg viewBox="0 0 24 24" fill="none" stroke="var(--cyan)" strokeWidth="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
            </svg>
            <span className="panel-title">Shared Storage Explorer</span>
          </div>
          <div style={{ fontSize: '12px', color: 'var(--cyan)', fontFamily: 'var(--font-mono)', marginTop: '6px' }}>
            {sharedDir || '~/SynX'}
          </div>
        </div>

        <div style={{ display: 'flex', gap: '10px' }}>
          <button className="btn btn-sm" onClick={onOpenFolder}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
              <polyline points="15 3 21 3 21 9" />
              <line x1="10" y1="14" x2="21" y2="3" />
            </svg>
            Open in OS
          </button>
          <button className="btn btn-sm btn-primary" onClick={loadFiles} disabled={loading}>
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="23 4 23 10 17 10" />
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
            </svg>
            {loading ? 'Refreshing…' : 'Refresh'}
          </button>
        </div>
      </div>

      {files.length === 0 ? (
        <div className="empty-placeholder">
          Shared folder is empty. Drop files on the Overview tab or copy files into your SynX folder.
        </div>
      ) : (
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th>Size</th>
              <th>Modified</th>
              <th style={{ textAlign: 'right' }}>Actions</th>
            </tr>
          </thead>
          <tbody>
            {files.map((f) => (
              <tr key={f.path || f.name}>
                <td>
                  <div className="file-cell">
                    {getFileTypeIcon(f.name)}
                    <span>{f.name}</span>
                  </div>
                </td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: '12px' }}>
                  {f.dir ? 'Folder' : formatBytes(f.size)}
                </td>
                <td style={{ fontFamily: 'var(--font-mono)', fontSize: '12px', color: 'var(--text-med)' }}>
                  {new Date(f.modified).toLocaleString()}
                </td>
                <td style={{ textAlign: 'right' }}>
                  <button
                    className="btn btn-sm btn-primary"
                    style={{ marginRight: '8px' }}
                    onClick={() => onSendFile(f.path || f.name)}
                  >
                    Send
                  </button>
                  <button
                    className="btn btn-sm"
                    style={{ color: 'var(--coral)', borderColor: 'rgba(255,45,96,0.3)' }}
                    onClick={() => handleDelete(f.path || f.name)}
                  >
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
};
