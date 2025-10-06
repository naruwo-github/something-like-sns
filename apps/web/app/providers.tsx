"use client";

import { Auth0Provider } from "@auth0/auth0-react";

export function Providers({ children }: { children: React.ReactNode }) {
  const domain = process.env.NEXT_PUBLIC_AUTH0_DOMAIN;
  const clientId = process.env.NEXT_PUBLIC_AUTH0_CLIENT_ID;

  if (!(domain && clientId)) {
    // In a real app, you'd want to show a proper error message.
    console.error(
      "Auth0 domain or client ID not set. Please check your .env.local file.",
    );
    return <>{children}</>;
  }

  return (
    <Auth0Provider
      domain={domain}
      clientId={clientId}
      authorizationParams={{
        redirect_uri:
          typeof window !== "undefined" ? window.location.origin : undefined,
      }}
    >
      {children}
    </Auth0Provider>
  );
}
