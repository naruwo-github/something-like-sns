import "server-only";

export const API_BASE =
  process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

type HeadersInput = {
  tenant: string;
  user: string;
};
export function createHeaders(
  input?: Partial<HeadersInput>,
): Record<string, string> {
  const tenant = input?.tenant ?? "acme";
  const user = input?.user ?? "u_alice";
  return { "X-Tenant": tenant, "X-User": user };
}

export async function rpc<TReq extends object, TRes>(
  path: string,
  req: TReq,
  headers: Record<string, string>,
): Promise<TRes> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    cache: "no-store",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      "Connect-Protocol-Version": "1",
      ...headers,
    },
    body: JSON.stringify(req ?? {}),
  });
  if (!res.ok) {
    let bodyText = "";
    try {
      bodyText = await res.text();
    } catch {}
    // eslint-disable-next-line no-console
    console.error(`RPC error ${res.status} for ${path}:`, bodyText);
    throw new Error(`HTTP ${res.status}`);
  }
  return res.json() as Promise<TRes>;
}
