'use client';

import React, { useState, useEffect, useRef } from 'react';
import { synxClient } from '@/lib/synx-client';

interface TerminalViewProps {
  nodeName: string;
}

export const TerminalView: React.FC<TerminalViewProps> = ({ nodeName }) => {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [output, setOutput] = useState<string>('SynX Developer Shell Ready.\nPress "Launch Terminal Session" to open interactive shell.');
  const [inputVal, setInputVal] = useState<string>('');
  const [history, setHistory] = useState<string[]>([]);
  const [histIndex, setHistIndex] = useState<number>(-1);
  const [loading, setLoading] = useState<boolean>(false);
  const outputEndRef = useRef<HTMLDivElement>(null);
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

  const startSession = async () => {
    setLoading(true);
    try {
      const sess = await synxClient.openTerminal(100, 30);
      if (sess && sess.session_id) {
        setSessionId(sess.session_id);
        setOutput((prev) => prev + `\n[Session Connected: ${sess.session_id}]\n`);
      }
    } finally {
      setLoading(false);
    }
  };

  const stopSession = async () => {
    if (sessionId) {
      await synxClient.closeTerminal(sessionId);
      setSessionId(null);
      setOutput((prev) => prev + '\n[Session Terminated]\n');
    }
  };

  // Poll terminal output if session is active
  useEffect(() => {
    if (!sessionId) return;

    pollIntervalRef.current = setInterval(async () => {
      const res = await synxClient.getTerminalOutput(sessionId);
      if (res.output) {
        setOutput((prev) => prev + res.output);
      }
      if (res.closed) {
        setSessionId(null);
        setOutput((prev) => prev + '\n[Process Exited]\n');
      }
    }, 400);

    return () => {
      if (pollIntervalRef.current) clearInterval(pollIntervalRef.current);
    };
  }, [sessionId]);

  // Auto-scroll to bottom of terminal
  useEffect(() => {
    outputEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [output]);

  const handleSend = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!inputVal.trim() && !inputVal.includes('\n')) return;

    const cmd = inputVal;
    setInputVal('');
    setHistory((prev) => [...prev, cmd]);
    setHistIndex(-1);

    if (sessionId) {
      await synxClient.sendTerminalInput(sessionId, cmd + '\n');
    } else {
      // Direct command execution fallback
      setOutput((prev) => prev + `\n> ${cmd}\nRunning...`);
      const res = await synxClient.executeCommand(cmd);
      if (res) {
        const out = (res.stdout || '') + (res.stderr ? `\nERR: ${res.stderr}` : '');
        setOutput((prev) => prev + `\n${out}\n[Exit: ${res.exit_code} in ${res.duration_ms}ms]\n`);
      }
    }
  };

  const runQuickCommand = async (cmd: string) => {
    if (sessionId) {
      await synxClient.sendTerminalInput(sessionId, cmd + '\n');
    } else {
      setOutput((prev) => prev + `\n> ${cmd}\nRunning...`);
      const res = await synxClient.executeCommand(cmd);
      if (res) {
        const out = (res.stdout || '') + (res.stderr ? `\nERR: ${res.stderr}` : '');
        setOutput((prev) => prev + `\n${out}\n[Exit: ${res.exit_code} in ${res.duration_ms}ms]\n`);
      }
    }
  };

  const clearOutput = () => {
    setOutput('SynX Developer Shell (cleared)\n');
  };

  return (
    <div className="page-fade-in" style={{ display: 'flex', flexDirection: 'column', height: '100%' }}>
      {/* Header bar */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px' }}>
        <div>
          <h3 style={{ fontFamily: 'var(--font-display)', fontSize: '20px', color: 'var(--text-pure)', fontWeight: 800 }}>
            Interactive Remote Terminal
          </h3>
          <p style={{ fontSize: '13px', color: 'var(--text-med)', marginTop: '2px' }}>
            Bi-directional shell streaming to {nodeName} via SynX protocol.
          </p>
        </div>

        <div style={{ display: 'flex', gap: '10px' }}>
          <button className="btn btn-sm" onClick={clearOutput}>
            Clear
          </button>
          {sessionId ? (
            <button className="btn btn-sm" style={{ borderColor: 'var(--coral)', color: 'var(--coral)' }} onClick={stopSession}>
              Disconnect Shell
            </button>
          ) : (
            <button className="btn btn-sm btn-primary" onClick={startSession} disabled={loading}>
              {loading ? 'Launching…' : '⚡ Launch Interactive Session'}
            </button>
          )}
        </div>
      </div>

      {/* Quick Actions Bar */}
      <div style={{ display: 'flex', gap: '8px', marginBottom: '12px', flexWrap: 'wrap', alignItems: 'center' }}>
        <span style={{ fontSize: '12px', fontWeight: 700, color: 'var(--text-dim)', marginRight: '4px' }}>Quick:</span>
        {['git status', 'docker ps', 'hostname', 'whoami', 'go version'].map((q) => (
          <button
            key={q}
            onClick={() => runQuickCommand(q)}
            style={{
              background: 'var(--bg-glass-pill)',
              border: '1px solid var(--border-dim)',
              color: 'var(--cyan)',
              fontFamily: 'var(--font-mono)',
              fontSize: '11px',
              padding: '4px 10px',
              borderRadius: '6px',
              cursor: 'pointer'
            }}
          >
            {q}
          </button>
        ))}
      </div>

      {/* Terminal Display */}
      <div
        style={{
          flex: 1,
          minHeight: '440px',
          background: '#040711',
          border: '1px solid var(--border-light)',
          borderRadius: 'var(--radius-md)',
          padding: '16px 20px',
          fontFamily: 'var(--font-mono)',
          fontSize: '13px',
          lineHeight: '1.5',
          color: '#38ef7d',
          overflowY: 'auto',
          boxShadow: 'inset 0 2px 10px rgba(0,0,0,0.8)'
        }}
      >
        <pre style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all', margin: 0 }}>
          {output}
        </pre>
        <div ref={outputEndRef} />
      </div>

      {/* Input Prompt */}
      <form onSubmit={handleSend} style={{ display: 'flex', gap: '10px', marginTop: '12px' }}>
        <div
          style={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            background: 'var(--bg-card)',
            border: '1px solid var(--border-focus)',
            borderRadius: 'var(--radius-sm)',
            padding: '0 12px'
          }}
        >
          <span style={{ color: 'var(--cyan)', fontFamily: 'var(--font-mono)', marginRight: '8px', fontWeight: 700 }}>
            {sessionId ? 'shell>' : 'cmd>'}
          </span>
          <input
            type="text"
            value={inputVal}
            onChange={(e) => setInputVal(e.target.value)}
            placeholder={sessionId ? 'Type keystrokes or shell command...' : 'Enter command to execute (e.g. git status, dir, node -v)...'}
            style={{
              flex: 1,
              background: 'transparent',
              border: 'none',
              outline: 'none',
              color: 'var(--text-pure)',
              fontFamily: 'var(--font-mono)',
              fontSize: '13px',
              padding: '10px 0'
            }}
          />
        </div>
        <button type="submit" className="btn btn-primary" style={{ padding: '0 20px' }}>
          Execute
        </button>
      </form>
    </div>
  );
};
