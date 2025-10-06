import AuthenticationButton from "./_components/AuthenticationButton";
import { Providers } from "./providers";

export const metadata = { title: "SNS", description: "Multi-tenant SNS" };

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="ja">
      <body>
        <Providers>
          <div style={{ maxWidth: 720, margin: "0 auto", padding: 16 }}>
            <nav
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                marginBottom: 16,
              }}
            >
              <div style={{ display: "flex", gap: 12 }}>
                <a href="/">Home</a>
                <a href="/dm">DM</a>
              </div>
              <AuthenticationButton />
            </nav>
            {children}
          </div>
        </Providers>
      </body>
    </html>
  );
}
