"use client";

import { useAuth0 } from "@auth0/auth0-react";

const AuthenticationButton = () => {
  const { isAuthenticated, user, loginWithRedirect, logout } = useAuth0();

  return isAuthenticated && user ? (
    <div style={{ display: "flex", gap: 8, alignItems: "center" }}>
      <span>{user.name}</span>
      <button
        type="button"
        onClick={() =>
          logout({ logoutParams: { returnTo: window.location.origin } })
        }
      >
        Logout
      </button>
    </div>
  ) : (
    <button type="button" onClick={() => loginWithRedirect()}>
      Login
    </button>
  );
};

export default AuthenticationButton;
