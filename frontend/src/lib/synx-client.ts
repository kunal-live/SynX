import { SynXState, SharedFile, TransferHistory } from '@/types/synx';

declare global {
  interface Window {
    go?: {
      app?: {
        App?: any;
      };
      main?: {
        App?: any;
      };
    };
    runtime?: {
      EventsOn?: (event: string, callback: (...args: any[]) => void) => () => void;
      EventsOff?: (event: string) => void;
    };
    __SYNX_LOCAL__?: {
      dir?: string;
      token?: string;
    };
  }
}

function getWailsApp() {
  if (typeof window === 'undefined') return null;
  return window.go?.app?.App || window.go?.main?.App || null;
}

export const synxClient = {
  async getState(): Promise<SynXState> {
    const w = getWailsApp();
    if (w?.GetState) {
      return await w.GetState();
    }

    try {
      const token = typeof window !== 'undefined'
        ? (window.__SYNX_LOCAL__?.token || localStorage.getItem('synx_token') || '')
        : '';
      const res = await fetch('/api/state', {
        headers: token ? { 'X-SynX-Token': token } : {}
      });
      if (res.ok) {
        return await res.json();
      }
    } catch (_) {}

    return {
      name: 'SynX Node',
      version: '1.0.0',
      platform: 'Windows',
      sharedDir: '~/SynX',
      token: 'synx-token',
      address: typeof window !== 'undefined' ? window.location.host || '0.0.0.0:8787' : '0.0.0.0:8787',
      peers: [],
      transfers: [],
      metrics: { bytes_sent: 0, bytes_received: 0 }
    };
  },

  async chooseFolder(): Promise<string> {
    const w = getWailsApp();
    if (w?.ChooseFolder) {
      return await w.ChooseFolder();
    }
    return prompt('Enter shared folder path:', '~/SynX') || '';
  },

  async setSharedDir(dir: string): Promise<void> {
    const w = getWailsApp();
    if (w?.SetSharedDir) {
      return await w.SetSharedDir(dir);
    }
    await fetch('/api/set-dir', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ dir })
    });
  },

  async openSharedFolder(): Promise<void> {
    const w = getWailsApp();
    if (w?.OpenSharedFolder) {
      return await w.OpenSharedFolder();
    }
  },

  async pair(address: string, token: string): Promise<void> {
    const w = getWailsApp();
    if (w?.Pair) {
      return await w.Pair(address, token);
    }
    const res = await fetch('/api/pair-peer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ address, token })
    });
    if (!res.ok) throw new Error('Pairing failed');
  },

  async generatePIN(): Promise<string> {
    const w = getWailsApp();
    if (w?.GeneratePairingPIN) {
      return await w.GeneratePairingPIN();
    }
    try {
      const res = await fetch('/api/v1/pair/start');
      if (res.ok) {
        const d = await res.json();
        return d.pin;
      }
    } catch (_) {}
    return '842 193';
  },

  async sendFile(peerId: string, path: string, destRel: string): Promise<void> {
    const w = getWailsApp();
    if (w?.SendFile) {
      return await w.SendFile(peerId, path, destRel);
    }
    throw new Error('Native SendFile available in desktop app');
  },

  async cancelTransfer(id: string): Promise<void> {
    const w = getWailsApp();
    if (w?.CancelTransfer) {
      return await w.CancelTransfer(id);
    }
  },

  async getSharedFiles(rel = ''): Promise<SharedFile[]> {
    const w = getWailsApp();
    if (w?.GetSharedFiles) {
      return await w.GetSharedFiles(rel);
    }
    try {
      const res = await fetch('/api/files?path=' + encodeURIComponent(rel));
      if (res.ok) return await res.json();
    } catch (_) {}
    return [];
  },

  async deleteSharedFile(rel: string): Promise<void> {
    const w = getWailsApp();
    if (w?.DeleteSharedFile) {
      return await w.DeleteSharedFile(rel);
    }
    await fetch('/api/delete?path=' + encodeURIComponent(rel));
  },

  async getHistory(limit = 50): Promise<TransferHistory[]> {
    const w = getWailsApp();
    if (w?.GetHistory) {
      return await w.GetHistory(limit);
    }
    try {
      const res = await fetch('/api/v1/history');
      if (res.ok) return await res.json();
    } catch (_) {}
    return [];
  },

  async uploadFile(file: File, path: string): Promise<void> {
    await fetch('/api/upload?path=' + encodeURIComponent(path), {
      method: 'POST',
      body: file
    });
  },

  // --- Developer Platform APIs (Section 10 & 15) ---

  async openTerminal(cols = 80, rows = 24, shell = ''): Promise<{ session_id: string } | null> {
    try {
      const res = await fetch('/api/terminal/open', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ cols, rows, shell })
      });
      if (res.ok) {
        const d = await res.json();
        return d.data || d;
      }
    } catch (e) {
      console.error('Failed to open terminal', e);
    }
    return null;
  },

  async sendTerminalInput(sessionId: string, data: string): Promise<void> {
    try {
      await fetch('/api/terminal/input', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: sessionId, data })
      });
    } catch (_) {}
  },

  async getTerminalOutput(sessionId: string): Promise<{ output: string; closed: boolean }> {
    try {
      const res = await fetch(`/api/terminal/output?session_id=${encodeURIComponent(sessionId)}`);
      if (res.ok) return await res.json();
    } catch (_) {}
    return { output: '', closed: false };
  },

  async closeTerminal(sessionId: string): Promise<void> {
    try {
      await fetch('/api/terminal/close', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ session_id: sessionId })
      });
    } catch (_) {}
  },

  async listTerminalSessions(): Promise<any[]> {
    try {
      const res = await fetch('/api/terminal/list');
      if (res.ok) return await res.json();
    } catch (_) {}
    return [];
  },

  async executeCommand(command: string, dir = '', timeout = 30): Promise<any> {
    try {
      const res = await fetch('/api/command/execute', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ command, dir, timeout })
      });
      if (res.ok) {
        const d = await res.json();
        return d.data || d;
      }
    } catch (e) {
      console.error('Failed to execute command', e);
    }
    return null;
  },

  async getClipboard(): Promise<string> {
    try {
      const res = await fetch('/api/clipboard');
      if (res.ok) {
        const d = await res.json();
        return d.data?.text || '';
      }
    } catch (_) {}
    return '';
  },

  async setClipboard(text: string): Promise<void> {
    try {
      await fetch('/api/clipboard', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ text })
      });
    } catch (_) {}
  },

  async getDevices(): Promise<any[]> {
    try {
      const res = await fetch('/api/devices');
      if (res.ok) {
        const d = await res.json();
        return d.devices || [];
      }
    } catch (_) {}
    return [];
  }
};
