import type { Metadata } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: 'SynX — Developer Connectivity Platform',
  description: 'Developer-first local network mesh for zero-config LAN discovery, interactive remote terminals, capability execution, and encrypted data streaming.',
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      <head>
        <meta name="color-scheme" content="dark" />
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
