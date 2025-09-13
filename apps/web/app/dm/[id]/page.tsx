"use client";
import { useEffect, useRef, useState } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_BASE || "http://localhost:8080";

type Message = {
  id: number;
  conversationId: number;
  senderUserId: number;
  body: string;
  createdAt: string;
};

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

export default function Conversation({ params }: { params: { id: string } }) {
  const cnvId = Number(params.id);
  const headers = { "X-Tenant": "acme", "X-User": "u_alice" };
  const [items, setItems] = useState<Message[]>([]);
  const [body, setBody] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    (async () => {
      await call("/sns.v1.TenantService/GetMe", {}, headers);
      const r = await call<{ conversation_id: number }, { items: Message[] }>(
        "/sns.v1.DMService/ListMessages",
        { conversation_id: cnvId },
        headers,
      );
      setItems(r.items);
    })();
  }, [cnvId]);

  const send = async () => {
    const r = await call<
      { conversation_id: number; body: string },
      { message: Message }
    >(
      "/sns.v1.DMService/SendMessage",
      { conversation_id: cnvId, body },
      headers,
    );
    setItems([...items, r.message]);
    setBody("");
    listRef.current?.scrollTo({ top: 1e9, behavior: "smooth" });
  };

  return (
    <main>
      <h1>会話 #{cnvId}</h1>
      <div
        ref={listRef}
        style={{
          display: "flex",
          flexDirection: "column",
          gap: 8,
          maxHeight: 400,
          overflow: "auto",
          border: "1px solid #ddd",
          padding: 8,
          marginBottom: 12,
        }}
      >
        {items.map((m) => (
          <div
            key={m.id}
            style={{
              alignSelf: m.senderUserId === 1 ? "flex-end" : "flex-start",
              background: "#f7f7f7",
              padding: 8,
              borderRadius: 8,
            }}
          >
            <div style={{ fontSize: 12, color: "#666" }}>
              {m.senderUserId} at {new Date(m.createdAt).toLocaleString()}
            </div>
            <div>{m.body}</div>
          </div>
        ))}
      </div>
      <div style={{ display: "flex", gap: 8 }}>
        <input
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="メッセージ"
          style={{ flex: 1, padding: 8 }}
        />
        <button type="button" onClick={send} disabled={!body.trim()}>
          送信
        </button>
      </div>
    </main>
  );
}
