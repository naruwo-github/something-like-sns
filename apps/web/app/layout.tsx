export const metadata = { title: 'SNS', description: 'Multi-tenant SNS' };

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>
        <div style={{ maxWidth: 720, margin: '0 auto', padding: 16 }}>
          {children}
        </div>
      </body>
    </html>
  );
}

