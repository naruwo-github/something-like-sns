export const metadata = { title: 'SNS', description: 'Multi-tenant SNS' };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>
        <div style={{ maxWidth: 720, margin: '0 auto', padding: 16 }}>
          <nav style={{ display: 'flex', gap: 12, marginBottom: 16 }}>
            <a href="/">Home</a>
            <a href="/dm">DM</a>
          </nav>
          {children}
        </div>
      </body>
    </html>
  );
}
