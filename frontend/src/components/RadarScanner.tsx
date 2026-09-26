'use client';

import React, { useEffect, useRef } from 'react';
import { Peer } from '@/types/synx';

interface RadarScannerProps {
  peers: Peer[];
  localName?: string;
}

export const RadarScanner: React.FC<RadarScannerProps> = ({ peers, localName = 'SynX' }) => {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);

  useEffect(() => {
    const canvas = canvasRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;

    let animId: number;
    let angle = 0;

    const resize = () => {
      const rect = canvas.getBoundingClientRect();
      canvas.width = rect.width * (window.devicePixelRatio || 1);
      canvas.height = rect.height * (window.devicePixelRatio || 1);
      ctx.scale(window.devicePixelRatio || 1, window.devicePixelRatio || 1);
    };

    resize();
    window.addEventListener('resize', resize);

    const render = () => {
      const rect = canvas.getBoundingClientRect();
      const cx = rect.width / 2;
      const cy = rect.height / 2;
      const maxR = Math.min(cx, cy) - 10;

      ctx.clearRect(0, 0, rect.width, rect.height);

      // Concentric Radar Rings
      [0.3, 0.6, 0.9].forEach(frac => {
        ctx.beginPath();
        ctx.arc(cx, cy, maxR * frac, 0, Math.PI * 2);
        ctx.strokeStyle = 'rgba(0, 240, 255, 0.16)';
        ctx.lineWidth = 1;
        ctx.stroke();
      });

      // Crosshairs
      ctx.beginPath();
      ctx.moveTo(cx - maxR, cy);
      ctx.lineTo(cx + maxR, cy);
      ctx.moveTo(cx, cy - maxR);
      ctx.lineTo(cx, cy + maxR);
      ctx.strokeStyle = 'rgba(0, 240, 255, 0.1)';
      ctx.lineWidth = 1;
      ctx.stroke();

      // Rotating Scanner Cone
      angle += 0.035;
      ctx.save();
      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.arc(cx, cy, maxR * 0.95, angle - 0.5, angle);
      ctx.closePath();
      const grad = ctx.createRadialGradient(cx, cy, 0, cx, cy, maxR);
      grad.addColorStop(0, 'rgba(0, 240, 255, 0.35)');
      grad.addColorStop(1, 'rgba(0, 240, 255, 0)');
      ctx.fillStyle = grad;
      ctx.fill();

      // Leading beam line
      ctx.beginPath();
      ctx.moveTo(cx, cy);
      ctx.lineTo(cx + Math.cos(angle) * maxR * 0.95, cy + Math.sin(angle) * maxR * 0.95);
      ctx.strokeStyle = 'rgba(0, 240, 255, 0.85)';
      ctx.lineWidth = 1.6;
      ctx.stroke();
      ctx.restore();

      // Center Node (Local Machine)
      ctx.beginPath();
      ctx.arc(cx, cy, 6, 0, Math.PI * 2);
      ctx.fillStyle = '#00f0ff';
      ctx.fill();
      ctx.beginPath();
      ctx.arc(cx, cy, 12, 0, Math.PI * 2);
      ctx.strokeStyle = 'rgba(0, 240, 255, 0.4)';
      ctx.lineWidth = 1.5;
      ctx.stroke();

      // Render Discovered Peers
      peers.forEach((p, idx) => {
        const pAngle = (idx * (Math.PI * 2 / Math.max(1, peers.length))) + 0.65;
        const pDist = maxR * 0.65;
        const px = cx + Math.cos(pAngle) * pDist;
        const py = cy + Math.sin(pAngle) * pDist;

        // Dotted Link
        ctx.beginPath();
        ctx.moveTo(cx, cy);
        ctx.lineTo(px, py);
        ctx.strokeStyle = 'rgba(0, 255, 157, 0.3)';
        ctx.lineWidth = 1;
        ctx.setLineDash([4, 4]);
        ctx.stroke();
        ctx.setLineDash([]);

        // Peer blip
        ctx.beginPath();
        ctx.arc(px, py, 5, 0, Math.PI * 2);
        ctx.fillStyle = '#00ff9d';
        ctx.fill();
        ctx.beginPath();
        ctx.arc(px, py, 11, 0, Math.PI * 2);
        ctx.strokeStyle = 'rgba(0, 255, 157, 0.5)';
        ctx.lineWidth = 1.5;
        ctx.stroke();

        // Label
        ctx.font = '10px "JetBrains Mono", monospace';
        ctx.fillStyle = '#f1f5f9';
        ctx.fillText(p.name || 'Peer', px + 12, py + 4);
      });

      animId = requestAnimationFrame(render);
    };

    animId = requestAnimationFrame(render);

    return () => {
      cancelAnimationFrame(animId);
      window.removeEventListener('resize', resize);
    };
  }, [peers]);

  return (
    <div className="radar-wrapper" title="Real-time LAN peer sonar scanner">
      <div className="radar-badge">
        <span className="pulse-beacon" style={{ background: 'var(--emerald)' }}></span>
        RADAR ACTIVE • UDP 8788
      </div>
      <canvas ref={canvasRef} id="radarCanvas" />
    </div>
  );
};
