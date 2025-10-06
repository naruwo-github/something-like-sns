export const metadata = { title: "SNS", description: "Multi-tenant SNS" };

import Providers from "./providers";
import { auth0 } from "../lib/auth0";

export default async function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const session = await auth0.getSession();
  const user = session?.user;
  return (
    <html lang="ja">
      <body>
        <div style={{ maxWidth: 720, margin: "0 auto", padding: 16 }}>
          <nav style={{ display: "flex", gap: 12, marginBottom: 16 }}>
            <a href="/">Home</a>
            <a href="/dm">DM</a>
            {user ? (
              <>
                <a href="/profile">Profile</a>
                <a href="/auth/logout">Logout</a>
              </>
            ) : (
              <a href="/auth/login">Login</a>
            )}
          </nav>
          <Providers>{children}</Providers>
        </div>
      </body>
    </html>
  );
}
