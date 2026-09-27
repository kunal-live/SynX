export interface Peer {
  id: string;
  device_id?: string;
  name: string;
  device_name?: string;
  address: string;
  port?: number;
  platform?: string;
  version?: string;
  capabilities?: string[];
  public_key?: string;
  trusted?: boolean;
  status?: string;
  last_seen?: number;
}

export interface TerminalSession {
  session_id: string;
  cols: number;
  rows: number;
  created_at: number;
  shell?: string;
}

export interface Transfer {
  id: string;
  name: string;
  size: number;
  done: number;
  progress: number;
  speed: string;
  eta: number;
  status: string;
  peer_id?: string;
  direction?: 'send' | 'receive';
}

export interface TransferHistory {
  id: string;
  file_name: string;
  size: number;
  direction: string;
  status: string;
  peer_id: string;
  completed_at: number;
}

export interface SharedFile {
  name: string;
  path: string;
  size: number;
  dir: boolean;
  modified: string;
}

export interface Metrics {
  bytes_sent?: number;
  bytes_received?: number;
  active_transfers?: number;
  connected_peers?: number;
}

export interface SynXState {
  name: string;
  version: string;
  platform: string;
  sharedDir: string;
  token: string;
  address: string;
  peers: Peer[];
  transfers: Transfer[];
  metrics: Metrics;
}
