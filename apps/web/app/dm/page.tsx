"use client";
import { useEffect, useState } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

type Conversation = { id: number; createdAt: string; memberUserIds: number[] };

async function call<TReq extends object, TRes>(
  path: string,
  req: TReq,
  headers: Record<string, string>,
): Promise<TRes> {
  const res = await fetch(`${API_BASE}${path}`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      Accept: "application/json",
      "Connect-Protocol-Version": "1",
      ...headers,
    },
    body: JSON.stringify(req ?? {}),
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json() as Promise<TRes>;
}

export default function DMList() {
  const headers = { "X-Tenant": "acme", "X-User": "u_alice" };
  const [items, setItems] = useState<Conversation[]>([]);
  const [other, setOther] = useState("");

  useEffect(() => {
    (async () => {
      await call("/sns.v1.TenantService/GetMe", {}, headers);
      const r = await call<{}, { items: Conversation[] }>(
        "/sns.v1.DMService/ListConversations",
        {},
        headers,
      );
      setItems(r.items);
    })();
  }, []);

  const createDM = async () => {
    const otherId = Number(other);
    if (!otherId) return;
    const r = await call<
      { other_user_id: number },
      { conversation_id: number }
    >("/sns.v1.DMService/GetOrCreateDM", { other_user_id: otherId }, headers);
    location.href = `/dm/${r.conversation_id}`;
  };

  return (
    <main>
      <h1>DM</h1>
      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input
          value={other}
          onChange={(e) => setOther(e.target.value)}
          placeholder="相手ユーザーID"
        />
        <button type="button" onClick={createDM} disabled={!other.trim()}>
          作成/開く
        </button>
      </div>
      <ul style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {items.map((c) => (
          <li key={c.id}>
            <a href={`/dm/${c.id}`}>
              会話 #{c.id} ({c.memberUserIds.join(", ")})
            </a>
          </li>
        ))}
      </ul>
    </main>
  );
}
