"use client";
import type { Message } from "../../_lib/types";
import { useEffect, useMemo, useRef, useState } from "react";
import { createHeaders, rpc } from "../../_lib/api";

type Props = {
  conversationId: number;
  initialItems: Message[];
};

export default function ConversationView({
  conversationId,
  initialItems,
}: Props) {
  const [items, setItems] = useState<Message[]>(initialItems);
  const [body, setBody] = useState("");
  const listRef = useRef<HTMLDivElement>(null);
  const lastMessageId = useMemo(
    () => (items.length ? items[items.length - 1].id : undefined),
    [items],
  );

  useEffect(() => {
    if (lastMessageId === undefined) return;
    listRef.current?.scrollTo({ top: 1e9, behavior: "smooth" });
  }, [lastMessageId]);

  const onSend = async () => {
    const r = await rpc<
      { conversation_id: number; body: string },
      { message: Message }
    >(
      "/sns.v1.DMService/SendMessage",
      { conversation_id: conversationId, body },
      createHeaders(),
    );
    setItems((prev) => [...prev, r.message]);
    setBody("");
  };

  return (
    <main>
      <h1>会話 #{conversationId}</h1>
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
        <button type="button" onClick={onSend} disabled={!body.trim()}>
          送信
        </button>
      </div>
    </main>
  );
}
