import NextAuth from "next-auth";
import Keycloak from "next-auth/providers/keycloak";

function decodeJwtExpiry(accessToken?: string) {
  if (!accessToken) {
    return undefined;
  }

  try {
    const [, payload] = accessToken.split(".");
    if (!payload) {
      return undefined;
    }

    const parsed = JSON.parse(
      Buffer.from(payload, "base64url").toString("utf-8")
    ) as { exp?: number };

    return parsed.exp ? parsed.exp * 1000 : undefined;
  } catch {
    return undefined;
  }
}

async function refreshAccessToken(token: Record<string, unknown>) {
  const issuer =
    process.env.KEYCLOAK_INTERNAL_ISSUER ?? process.env.KEYCLOAK_ISSUER;
  const clientId = process.env.KEYCLOAK_CLIENT_ID;
  const clientSecret = process.env.KEYCLOAK_CLIENT_SECRET;

  if (!issuer || !clientId || !clientSecret || !token.refreshToken) {
    return { ...token, error: "RefreshAccessTokenError" as const };
  }

  const refreshTokenExpires = Number(token.refreshTokenExpires ?? 0);
  if (refreshTokenExpires > 0 && Date.now() >= refreshTokenExpires - 10_000) {
    return {
      ...token,
      accessToken: undefined,
      accessTokenExpires: 0,
      refreshToken: undefined,
      refreshTokenExpires: 0,
      error: "RefreshTokenExpired" as const,
    };
  }

  try {
    const tokenEndpoint = new URL(
      "protocol/openid-connect/token",
      issuer.endsWith("/") ? issuer : `${issuer}/`
    );

    const body = new URLSearchParams({
      grant_type: "refresh_token",
      refresh_token: String(token.refreshToken),
      client_id: clientId,
      client_secret: clientSecret,
    });

    const response = await fetch(tokenEndpoint.toString(), {
      method: "POST",
      headers: { "Content-Type": "application/x-www-form-urlencoded" },
      body,
    });

    if (!response.ok) {
      throw new Error("Failed to refresh token");
    }

    const refreshed = (await response.json()) as {
      access_token: string;
      expires_in: number;
      refresh_token?: string;
      refresh_expires_in?: number;
      id_token?: string;
    };

    return {
      ...token,
      accessToken: refreshed.access_token,
      accessTokenExpires: Date.now() + refreshed.expires_in * 1000,
      refreshToken: refreshed.refresh_token ?? token.refreshToken,
      refreshTokenExpires: refreshed.refresh_expires_in
        ? Date.now() + refreshed.refresh_expires_in * 1000
        : token.refreshTokenExpires,
      idToken: refreshed.id_token ?? token.idToken,
      error: undefined,
    };
  } catch {
    return {
      ...token,
      accessToken: undefined,
      accessTokenExpires: 0,
      refreshToken: undefined,
      refreshTokenExpires: 0,
      error: "RefreshAccessTokenError" as const,
    };
  }
}

export const { handlers, auth, signIn, signOut } = NextAuth({
  secret: process.env.AUTH_SECRET ?? process.env.NEXTAUTH_SECRET,
  providers: [
    (() => {
      const publicIssuer = process.env.KEYCLOAK_ISSUER;
      const internalIssuer =
        process.env.KEYCLOAK_INTERNAL_ISSUER ?? publicIssuer;

      const internalBase = internalIssuer?.endsWith("/")
        ? internalIssuer.slice(0, -1)
        : internalIssuer;
      const publicBase = publicIssuer?.endsWith("/")
        ? publicIssuer.slice(0, -1)
        : publicIssuer;

      return Keycloak({
        clientId: process.env.KEYCLOAK_CLIENT_ID,
        clientSecret: process.env.KEYCLOAK_CLIENT_SECRET,
        issuer: publicIssuer,
        authorization: publicBase
          ? `${publicBase}/protocol/openid-connect/auth`
          : undefined,
        token: internalBase
          ? `${internalBase}/protocol/openid-connect/token`
          : undefined,
        userinfo: internalBase
          ? `${internalBase}/protocol/openid-connect/userinfo`
          : undefined,
      });
    })(),
  ],
  callbacks: {
    async jwt({ token, account }) {
      if (account) {
        token.accessToken = account.access_token;
        token.idToken = account.id_token;
        token.refreshToken = account.refresh_token;
        const expiresAtMs = Number(account.expires_at);
        const expiresInMs = Number(account.expires_in ?? 300) * 1000;
        token.accessTokenExpires = Number.isFinite(expiresAtMs)
          ? expiresAtMs * 1000
          : Date.now() + expiresInMs;
        token.refreshTokenExpires = account.refresh_expires_in
          ? Date.now() + Number(account.refresh_expires_in) * 1000
          : undefined;
        token.error = undefined;
        return token;
      }

      if (!token.accessTokenExpires) {
        token.accessTokenExpires = decodeJwtExpiry(
          token.accessToken as string | undefined
        );
      }

      const accessTokenExpires = Number(token.accessTokenExpires ?? 0);

      if (
        token.accessToken &&
        accessTokenExpires > 0 &&
        Date.now() < accessTokenExpires - 10_000
      ) {
        return token;
      }

      // If token expiry cannot be determined, avoid forced refresh loops.
      // In this case, the API 401 handler will redirect to login when needed.
      if (!accessTokenExpires) {
        return token;
      }

      return await refreshAccessToken(token);
    },
    async session({ session, token }) {
      if (token?.accessToken) {
        session.accessToken = token.accessToken as string;
      }
      if (token?.idToken) {
        session.idToken = token.idToken as string;
      }
      if (token?.error) {
        session.error = token.error as string;
      }
      return session;
    },
  },
});
