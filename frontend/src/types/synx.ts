export interface Peer {
  id: string;
  name: string;
  address: string;
  platform?: string;
  trusted?: boolean;
  status?: string;
  last_seen?: number;
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
