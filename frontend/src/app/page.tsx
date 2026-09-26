'use client';

import React, { useEffect, useState, useCallback, useRef } from 'react';
import { SynXState, Peer } from '@/types/synx';
import { synxClient } from '@/lib/synx-client';
import { Sidebar } from '@/components/Sidebar';
import { Header } from '@/components/Header';
import { RadarScanner } from '@/components/RadarScanner';
import { StatsGrid } from '@/components/StatsGrid';
import { DropZonePod } from '@/components/DropZonePod';
import { TransfersView } from '@/components/TransfersView';
import { DevicesView } from '@/components/DevicesView';
import { StorageExplorer } from '@/components/StorageExplorer';
import { SettingsView } from '@/components/SettingsView';
import { PairingModal } from '@/components/PairingModal';
import { Toast } from '@/components/Toast';

export default function Home() {
  const [state, setState] = useState<SynXState>({
    name: 'SynX Node',
    version: '1.0.0',
    platform: 'Windows',
    sharedDir: '~/SynX',
    token: '',
    address: '0.0.0.0:8787',
    peers: [],
    transfers: [],
    metrics: {}
  });

  const [currentPage, setCurrentPage] = useState<string>('overview');
  const [selectedTarget, setSelectedTarget] = useState<string>('all');
  const [isPairModalOpen, setIsPairModalOpen] = useState<boolean>(false);
  const [pinCode, setPinCode] = useState<string>('842 193');
  const [toastMessage, setToastMessage] = useState<string | null>(null);
  const toastTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const notify = useCallback((msg: string) => {
    if (toastTimeoutRef.current) clearTimeout(toastTimeoutRef.current);
    setToastMessage(msg);
    toastTimeoutRef.current = setTimeout(() => {
      setToastMessage(null);
    }, 3200);
  }, []);

  const refreshState = useCallback(async () => {
    try {
      const data = await synxClient.getState();
      if (data) setState(data);
    } catch (e) {
      console.error('Failed to fetch SynX state', e);
    }
  }, []);

  useEffect(() => {
    refreshState();
    const interval = setInterval(refreshState, 2500);

    // Wails runtime events
    const unsubDiscover = window.runtime?.EventsOn?.('peer.discovered', refreshState);
    const unsubDisconnect = window.runtime?.EventsOn?.('peer.disconnected', refreshState);
    const unsubTransferStart = window.runtime?.EventsOn?.('transfer.started', refreshState);
    const unsubTransferDone = window.runtime?.EventsOn?.('transfer.completed', () => {
      refreshState();
      notify('File transfer completed & verified!');
    });

    return () => {
      clearInterval(interval);
      unsubDiscover?.();
      unsubDisconnect?.();
      unsubTransferStart?.();
      unsubTransferDone?.();
    };
  }, [refreshState, notify]);

  // Actions
  const handleCopy = (text: string, label: string) => {
    if (!text) return;
    navigator.clipboard.writeText(text);
    notify(`Copied ${label}: ${text}`);
  };

  const handleOpenFolder = async () => {
    await synxClient.openSharedFolder();
  };

  const handleBrowseFolder = async () => {
    const dir = await synxClient.chooseFolder();
    if (dir) {
      await synxClient.setSharedDir(dir);
      notify(`Shared folder updated to ${dir}`);
      refreshState();
    }
  };

  const handleGeneratePIN = async () => {
    const pin = await synxClient.generatePIN();
    setPinCode(pin || '842 193');
  };

  const handleOpenPairModal = () => {
    setIsPairModalOpen(true);
    handleGeneratePIN();
  };

  const handleFilesSelected = async (files: File[]) => {
    if (!files.length) return;
    if (state.peers.length === 0) {
      notify(`${files.length} file(s) staged. Discover or pair a peer to send.`);
      return;
    }

    let targets: Peer[] = [];
    if (selectedTarget === 'all') {
      targets = state.peers;
    } else {
      const p = state.peers.find((x) => x.id === selectedTarget);
      targets = p ? [p] : [state.peers[0]];
    }

    notify(`Broadcasting ${files.length} file(s) to ${targets.length} peer(s)…`);

    for (const peer of targets) {
      for (const f of files) {
        try {
          const filePath = (f as any).path;
          if (filePath) {
            await synxClient.sendFile(peer.id, filePath, f.name);
          } else {
            await synxClient.uploadFile(f, f.name);
          }
        } catch (err: any) {
          console.error(err);
        }
      }
    }

    notify('Files queued in transfer stream');
    refreshState();
  };

  const handleQuickSendToPeer = (peerId: string) => {
    setSelectedTarget(peerId);
    setCurrentPage('overview');
    notify('Drop or pick files to send directly to selected peer');
  };

  const handleSendStorageFile = async (relPath: string) => {
    if (state.peers.length === 0) {
      notify('No connected peers on LAN to send to.');
      return;
    }
    const peer = state.peers[0];
    const fullPath = (state.sharedDir ? state.sharedDir + '/' : '') + relPath;
    try {
      await synxClient.sendFile(peer.id, fullPath, relPath);
      notify(`Transfer queued for ${peer.name}`);
      refreshState();
    } catch (err: any) {
      notify('Transfer error: ' + err.message);
    }
  };

  // Header Titles Mapping
  const titles: Record<string, [string, string]> = {
    overview: ['Mesh Overview', 'Real-time LAN peer topology and streaming queue'],
    devices: ['Connected Nodes', 'Discovered LAN devices and verified trust records'],
    transfers: ['Transfer Streams', 'High-throughput chunk streams and verified commits'],
    explorer: ['Shared Storage', 'Direct access to your local SynX shared directory'],
    settings: ['Node Settings', 'Security tokens, port allocation, and storage directories']
  };

  const [currentTitle, currentSubtitle] = titles[currentPage] || ['SynX', ''];

  return (
    <div className="app-shell">
      {/* Sidebar Navigation */}
      <Sidebar
        currentPage={currentPage}
        onPageChange={setCurrentPage}
        peerCount={state.peers.length}
        transferCount={state.transfers.length}
        nodeName={state.name}
        nodeAddress={state.address}
        onCopyAddress={() => handleCopy(state.address, 'address')}
        onOpenFolder={handleOpenFolder}
      />

      {/* Main Content Stage */}
      <main className="main-stage">
        <Header
          title={currentTitle}
          subtitle={currentSubtitle}
          peerCount={state.peers.length}
          onOpenPairModal={handleOpenPairModal}
          onBroadcastClick={() => setCurrentPage('overview')}
        />

        <section className="content-viewport">
          {currentPage === 'overview' && (
            <div className="page-fade-in">
              {/* Hero Mesh Card with Animated Radar */}
              <div className="hero-mesh-card">
                <div>
                  <div className="hero-info-tag">
                    <span className="pulse-beacon" style={{ background: 'var(--cyan)', boxShadow: '0 0 8px var(--cyan)' }}></span>
                    ⚡ Zero-Cloud • Decentralized P2P Mesh
                  </div>
                  <h2>Ultra-Fast Local Network File Transfer</h2>
                  <p>
                    SynX establishes direct, high-throughput peer-to-peer pipelines across all devices on your Wi-Fi and Ethernet. Zero external cloud servers, 8MB chunked streaming, and real-time SHA-256 integrity verification.
                  </p>
                  <div className="hero-actions">
                    <button className="btn btn-primary" onClick={() => setCurrentPage('devices')}>
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
                        <path d="M2 12h20" />
                      </svg>
                      Explore Mesh Map
                    </button>
                    <button className="btn" onClick={handleOpenFolder}>
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
                      </svg>
                      Open Local Folder
                    </button>
                  </div>
                </div>

                <RadarScanner peers={state.peers} localName={state.name} />
              </div>

              {/* Performance Stats Quad */}
              <StatsGrid
                peerCount={state.peers.length}
                transferCount={state.transfers.length}
                metrics={state.metrics}
              />

              {/* Holographic Drop Zone */}
              <DropZonePod
                peers={state.peers}
                selectedTarget={selectedTarget}
                onSelectTarget={setSelectedTarget}
                onFilesSelected={handleFilesSelected}
              />

              {/* Split Layout: Active Streams & Connected Nodes */}
              <div className="split-layout">
                <div className="panel-card">
                  <div className="panel-head">
                    <div className="panel-title-group">
                      <svg viewBox="0 0 24 24" fill="none" stroke="var(--cyan)" strokeWidth="2">
                        <polyline points="17 1 21 5 17 9" />
                        <path d="M3 11V9a4 4 0 0 1 4-4h14" />
                      </svg>
                      <span className="panel-title">Active Streams Queue</span>
                    </div>
                    <button className="btn btn-sm" onClick={() => setCurrentPage('transfers')}>
                      Full Center →
                    </button>
                  </div>
                  <TransfersView
                    transfers={state.transfers.slice(0, 3)}
                    onCancelTransfer={(id) => synxClient.cancelTransfer(id)}
                  />
                </div>

                <div className="panel-card">
                  <div className="panel-head">
                    <div className="panel-title-group">
                      <svg viewBox="0 0 24 24" fill="none" stroke="var(--emerald)" strokeWidth="2">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20" />
                      </svg>
                      <span className="panel-title">Nearby LAN Devices</span>
                    </div>
                    <button className="btn btn-sm" onClick={handleOpenPairModal}>
                      + Pair Node
                    </button>
                  </div>
                  <DevicesView
                    peers={state.peers.slice(0, 4)}
                    onOpenPairModal={handleOpenPairModal}
                    onShowPINModal={handleOpenPairModal}
                    onSendToPeer={handleQuickSendToPeer}
                  />
                </div>
              </div>
            </div>
          )}

          {currentPage === 'devices' && (
            <div className="page-fade-in">
              <DevicesView
                peers={state.peers}
                onOpenPairModal={handleOpenPairModal}
                onShowPINModal={handleOpenPairModal}
                onSendToPeer={handleQuickSendToPeer}
              />
            </div>
          )}

          {currentPage === 'transfers' && (
            <div className="page-fade-in">
              <TransfersView
                transfers={state.transfers}
                onCancelTransfer={(id) => synxClient.cancelTransfer(id)}
              />
            </div>
          )}

          {currentPage === 'explorer' && (
            <div className="page-fade-in">
              <StorageExplorer
                sharedDir={state.sharedDir}
                onOpenFolder={handleOpenFolder}
                onSendFile={handleSendStorageFile}
                onNotify={notify}
              />
            </div>
          )}

          {currentPage === 'settings' && (
            <div className="page-fade-in">
              <SettingsView
                sharedDir={state.sharedDir}
                nodeName={state.name}
                token={state.token}
                address={state.address}
                onBrowseFolder={handleBrowseFolder}
                onOpenFolder={handleOpenFolder}
                onCopyToken={() => handleCopy(state.token, 'security token')}
                onSavePreferences={() => notify('Configuration saved successfully')}
              />
            </div>
          )}
        </section>
      </main>

      {/* Device Pairing Modal */}
      <PairingModal
        isOpen={isPairModalOpen}
        pin={pinCode}
        onClose={() => setIsPairModalOpen(false)}
        onGeneratePIN={handleGeneratePIN}
        onCopyPIN={() => handleCopy(pinCode, 'PIN')}
        onSubmitDirectPair={async (addr, tok) => {
          await synxClient.pair(addr, tok);
          notify('Device paired and trusted successfully!');
          refreshState();
        }}
      />

      {/* Floating Action Toast */}
      <Toast message={toastMessage} />
    </div>
  );
}
