import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'SynX — Next-Gen Decentralized LAN File Transfer Mesh',
  description: 'Ultra-fast, zero-cloud peer-to-peer file streaming over local Wi-Fi and Ethernet. Powered by chunked parallel streams, real-time mesh discovery, and cryptographic SHA-256 verification.',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;600;700;800&family=Plus+Jakarta+Sans:wght@400;500;600;700;800&family=Space+Grotesk:wght@600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body>
        <div className="ambient-orb orb-1"></div>
        <div className="ambient-orb orb-2"></div>
        <div className="ambient-orb orb-3"></div>
        {children}
      </body>
    </html>
  );
}
