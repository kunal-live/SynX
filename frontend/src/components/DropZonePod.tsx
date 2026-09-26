'use client';

import React, { useRef, useState } from 'react';
import { Peer } from '@/types/synx';

interface DropZonePodProps {
  peers: Peer[];
  selectedTarget: string;
  onSelectTarget: (id: string) => void;
  onFilesSelected: (files: File[]) => void;
}

export const DropZonePod: React.FC<DropZonePodProps> = ({
  peers,
  selectedTarget,
  onSelectTarget,
  onFilesSelected
}) => {
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);

  const handleDragOver = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(true);
  };

  const handleDragLeave = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
  };

  const handleDrop = (e: React.DragEvent) => {
    e.preventDefault();
    setIsDragOver(false);
    if (e.dataTransfer.files && e.dataTransfer.files.length > 0) {
      onFilesSelected(Array.from(e.dataTransfer.files));
    }
  };

  const handleFileInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (e.target.files && e.target.files.length > 0) {
      onFilesSelected(Array.from(e.target.files));
    }
  };

  return (
    <div className="dropzone-container">
      <div
        id="dropZone"
        className={`drop-zone ${isDragOver ? 'dragover' : ''}`}
        onDragOver={handleDragOver}
        onDragLeave={handleDragLeave}
        onDrop={handleDrop}
        onClick={(e) => {
          if ((e.target as HTMLElement).closest('.target-peers-bar')) return;
          fileInputRef.current?.click();
        }}
      >
        <div className="laser-sweep" style={{ display: isDragOver ? 'block' : 'none' }}></div>
        <div className="drop-icon-sphere">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
            <polyline points="17 8 12 3 7 8" />
            <line x1="12" y1="3" x2="12" y2="15" />
          </svg>
        </div>
        <div className="drop-title">DRAG &amp; DROP FILES TO BROADCAST</div>
        <div className="drop-subtitle">
          Drop files anywhere inside this pod or click to pick from your storage. Files are automatically chunked and streamed directly to selected peers.
        </div>

        {/* Target Peer Selector Chips */}
        <div className="target-peers-bar" onClick={(e) => e.stopPropagation()}>
          <span style={{ fontSize: '12px', fontWeight: 700, color: 'var(--text-med)', marginRight: '4px' }}>
            Target:
          </span>
          <div
            className={`target-peer-chip ${selectedTarget === 'all' ? 'selected' : ''}`}
            onClick={() => onSelectTarget('all')}
          >
            ⚡ All Connected Nodes
          </div>
          {peers.map((p) => (
            <div
              key={p.id}
              className={`target-peer-chip ${selectedTarget === p.id ? 'selected' : ''}`}
              onClick={() => onSelectTarget(p.id)}
            >
              {p.name || 'Node'} ({p.address ? p.address.split(':')[0] : 'LAN'})
            </div>
          ))}
        </div>

        <input
          type="file"
          ref={fileInputRef}
          style={{ display: 'none' }}
          multiple
          onChange={handleFileInputChange}
        />
      </div>
    </div>
  );
};
