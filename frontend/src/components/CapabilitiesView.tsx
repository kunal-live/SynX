'use client';

import React, { useState, useEffect } from 'react';
import { synxClient } from '@/lib/synx-client';

export const CapabilitiesView: React.FC = () => {
  const [clipboardText, setClipboardText] = useState<string>('');
  const [newClipText, setNewClipText] = useState<string>('');
  const [commandInput, setCommandInput] = useState<string>('git status');
  const [commandResult, setCommandResult] = useState<any>(null);
  const [runningCmd, setRunningCmd] = useState<boolean>(false);
  const [statusMsg, setStatusMsg] = useState<string>('');

  const refreshClipboard = async () => {
    const text = await synxClient.getClipboard();
    setClipboardText(text);
  };

  useEffect(() => {
    refreshClipboard();
  }, []);

  const handleSetClipboard = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newClipText.trim()) return;
    await synxClient.setClipboard(newClipText);
    setStatusMsg('Clipboard updated!');
    setNewClipText('');
    refreshClipboard();
    setTimeout(() => setStatusMsg(''), 2500);
  };

  const handleRunCommand = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!commandInput.trim()) return;
    setRunningCmd(true);
    try {
      const res = await synxClient.executeCommand(commandInput);
      setCommandResult(res);
    } finally {
      setRunningCmd(false);
    }
  };

  const capabilities = [
    {
      name: 'Interactive Terminal',
      id: 'terminal',
      version: '1.0.0',
      status: 'Active',
      color: 'var(--cyan)',
      desc: 'Pseudo-terminal streaming with stdin/stdout piping, process signal handling, and window resize.',
      actions: ['terminal.open', 'terminal.input', 'terminal.resize', 'terminal.close']
    },
    {
      name: 'Remote Command Execution',
      id: 'command',
      version: '1.0.0',
      status: 'Active',
      color: 'var(--emerald)',
      desc: 'Audited, authenticated single-command execution with timeout enforcement and structured outputs.',
      actions: ['command.execute']
    },
    {
      name: 'Chunked File Streaming',
      id: 'files',
      version: '1.0.0',
      status: 'Active',
      color: 'var(--violet)',
      desc: 'High-speed 8MB chunked parallel streaming with streaming SHA-256 verification and atomic commit.',
      actions: ['files.negotiate', 'files.chunk', 'files.commit']
    },
    {
      name: 'Secure Clipboard Sync',
      id: 'clipboard',
      version: '1.0.0',
      status: 'Active',
      color: 'var(--amber)',
      desc: 'Explicit cross-workstation clipboard synchronization with opt-in permissions.',
      actions: ['clipboard.get', 'clipboard.set']
    },
    {
      name: 'Developer Automation & Events',
      id: 'events',
      version: '1.0.0',
      status: 'Active',
      color: 'var(--cyan)',
      desc: 'Pub/sub event bus emitting device.connected, build.completed, and transfer progress hooks.',
      actions: ['events.subscribe', 'events.publish']
    }
  ];

  return (
    <div className="page-fade-in">
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '24px' }}>
        <div>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '20px', color: 'var(--text-pure)', fontWeight: 800 }}>
            Capability Registry &amp; Tools
          </h3>
          <p style={{ fontSize: '13px', color: 'var(--text-med)', marginTop: '2px' }}>
            Inspect advertised capabilities, test developer tools, and manage execution policies.
          </p>
        </div>
      </div>

      {/* Capabilities Grid */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(320px, 1fr))', gap: '16px', marginBottom: '28px' }}>
        {capabilities.map((c) => (
          <div key={c.id} className="panel-card" style={{ padding: '20px' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '10px' }}>
              <span style={{ fontFamily: 'var(--font-display)', fontWeight: 700, fontSize: '16px', color: 'var(--text-pure)' }}>
                {c.name}
              </span>
              <span
                style={{
                  fontSize: '11px',
                  fontWeight: 700,
                  padding: '2px 8px',
                  borderRadius: '999px',
                  background: 'var(--bg-glass-pill)',
                  border: `1px solid ${c.color}`,
                  color: c.color
                }}
              >
                v{c.version} • {c.status}
              </span>
            </div>
            <p style={{ fontSize: '13px', color: 'var(--text-med)', lineHeight: '1.45', marginBottom: '14px' }}>
              {c.desc}
            </p>
            <div style={{ display: 'flex', gap: '6px', flexWrap: 'wrap' }}>
              {c.actions.map((act) => (
                <span
                  key={act}
                  style={{
                    fontFamily: 'var(--font-mono)',
                    fontSize: '10px',
                    padding: '2px 6px',
                    borderRadius: '4px',
                    background: 'rgba(255,255,255,0.04)',
                    color: 'var(--text-dim)'
                  }}
                >
                  {act}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Split Playground: Clipboard & Command Tester */}
      <div className="split-layout">
        {/* Clipboard Sync Pod */}
        <div className="panel-card">
          <div className="panel-head">
            <span className="panel-title">📋 Remote Clipboard Tool</span>
            <button className="btn btn-sm" onClick={refreshClipboard}>Refresh</button>
          </div>
          <div style={{ marginBottom: '16px' }}>
            <label style={{ fontSize: '12px', color: 'var(--text-dim)', display: 'block', marginBottom: '6px' }}>
              Current Shared Clipboard Content:
            </label>
            <div
              style={{
                background: 'var(--bg-void)',
                border: '1px solid var(--border-dim)',
                borderRadius: 'var(--radius-sm)',
                padding: '10px 14px',
                fontFamily: 'var(--font-mono)',
                fontSize: '13px',
                color: clipboardText ? 'var(--cyan)' : 'var(--text-dim)',
                minHeight: '44px',
                whiteSpace: 'pre-wrap'
              }}
            >
              {clipboardText || '(Empty clipboard buffer)'}
            </div>
          </div>
          <form onSubmit={handleSetClipboard} style={{ display: 'flex', gap: '8px' }}>
            <input
              type="text"
              value={newClipText}
              onChange={(e) => setNewClipText(e.target.value)}
              placeholder="Paste or type text to sync across mesh..."
              style={{
                flex: 1,
                background: 'var(--bg-void)',
                border: '1px solid var(--border-light)',
                borderRadius: 'var(--radius-sm)',
                padding: '8px 12px',
                color: 'var(--text-pure)',
                fontSize: '13px'
              }}
            />
            <button type="submit" className="btn btn-sm btn-primary">
              Sync Text
            </button>
          </form>
          {statusMsg && <div style={{ fontSize: '12px', color: 'var(--emerald)', marginTop: '8px' }}>{statusMsg}</div>}
        </div>

        {/* Command Runner Pod */}
        <div className="panel-card">
          <div className="panel-head">
            <span className="panel-title">⚡ Command Execution Tester</span>
          </div>
          <form onSubmit={handleRunCommand} style={{ display: 'flex', gap: '8px', marginBottom: '14px' }}>
            <input
              type="text"
              value={commandInput}
              onChange={(e) => setCommandInput(e.target.value)}
              placeholder="e.g. git status, hostname, dir..."
              style={{
                flex: 1,
                background: 'var(--bg-void)',
                border: '1px solid var(--border-light)',
                borderRadius: 'var(--radius-sm)',
                padding: '8px 12px',
                fontFamily: 'var(--font-mono)',
                color: 'var(--text-pure)',
                fontSize: '13px'
              }}
            />
            <button type="submit" className="btn btn-sm btn-primary" disabled={runningCmd}>
              {runningCmd ? 'Running…' : 'Run'}
            </button>
          </form>

          {commandResult && (
            <div
              style={{
                background: '#040711',
                border: '1px solid var(--border-dim)',
                borderRadius: 'var(--radius-sm)',
                padding: '12px',
                fontFamily: 'var(--font-mono)',
                fontSize: '12px',
                color: '#38ef7d',
                maxHeight: '180px',
                overflowY: 'auto'
              }}
            >
              <div style={{ color: 'var(--text-dim)', marginBottom: '4px' }}>
                Exit: {commandResult.exit_code} | Duration: {commandResult.duration_ms}ms
              </div>
              <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
                {commandResult.stdout || commandResult.stderr || '(no output)'}
              </pre>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
