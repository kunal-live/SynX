'use client';

import React, { useState } from 'react';

interface PairingModalProps {
  isOpen: boolean;
  pin: string;
  onClose: () => void;
  onGeneratePIN: () => void;
  onCopyPIN: () => void;
  onSubmitDirectPair: (address: string, token: string) => Promise<void>;
}

export const PairingModal: React.FC<PairingModalProps> = ({
  isOpen,
  pin,
  onClose,
  onGeneratePIN,
  onCopyPIN,
  onSubmitDirectPair
}) => {
  const [activeTab, setActiveTab] = useState<'pin' | 'direct'>('pin');
  const [addressInput, setAddressInput] = useState('');
  const [tokenInput, setTokenInput] = useState('');
  const [submitting, setSubmitting] = useState(false);

  if (!isOpen) return null;

  const handleDirectSubmit = async () => {
    if (!addressInput.trim() || !tokenInput.trim()) {
      alert('Please enter both address and token.');
      return;
    }
    setSubmitting(true);
    try {
      await onSubmitDirectPair(addressInput.trim(), tokenInput.trim());
      onClose();
    } catch (err: any) {
      alert('Pairing failed: ' + err.message);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="modal-overlay" style={{ display: 'grid' }}>
      <div className="modal-dialog">
        <div className="modal-header">
          <div className="modal-title">Device Pairing &amp; Mutual Trust</div>
          <button className="modal-close-btn" onClick={onClose}>✕</button>
        </div>

        <div style={{ display: 'flex', gap: '10px', borderBottom: '1px solid var(--border-dim)', paddingBottom: '14px' }}>
          <button
            className={`btn btn-sm ${activeTab === 'pin' ? 'btn-primary' : ''}`}
            onClick={() => setActiveTab('pin')}
          >
            6-Digit PIN Code
          </button>
          <button
            className={`btn btn-sm ${activeTab === 'direct' ? 'btn-primary' : ''}`}
            onClick={() => setActiveTab('direct')}
          >
            Direct LAN IP &amp; Token
          </button>
        </div>

        {activeTab === 'pin' ? (
          <div>
            <p style={{ fontSize: '13px', color: 'var(--text-med)', marginBottom: '16px', lineHeight: 1.5 }}>
              Input this one-time secure pairing PIN into any SynX peer node on your local network:
            </p>
            <div className="pin-hologram-box">
              <div className="neon-pin-text">{pin || '842 193'}</div>
              <div className="pin-countdown-text">Cryptographically verified • Expires in 5 minutes</div>
            </div>
            <div style={{ marginTop: '18px', display: 'flex', gap: '12px' }}>
              <button className="btn btn-primary" style={{ flex: 1 }} onClick={onGeneratePIN}>
                Generate New PIN
              </button>
              <button className="btn" onClick={onCopyPIN}>
                Copy Code
              </button>
            </div>
          </div>
        ) : (
          <div>
            <div className="field-block">
              <label className="field-label">Peer LAN Endpoint</label>
              <input
                className="input-control"
                placeholder="e.g. 192.168.1.15:8787"
                value={addressInput}
                onChange={(e) => setAddressInput(e.target.value)}
              />
            </div>
            <div className="field-block">
              <label className="field-label">Peer Security Token</label>
              <input
                className="input-control"
                placeholder="Paste remote node pairing token"
                value={tokenInput}
                onChange={(e) => setTokenInput(e.target.value)}
              />
            </div>
            <div style={{ marginTop: '18px' }}>
              <button
                className="btn btn-primary"
                style={{ width: '100%' }}
                onClick={handleDirectSubmit}
                disabled={submitting}
              >
                {submitting ? 'Connecting…' : 'Pair & Trust Endpoint'}
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
