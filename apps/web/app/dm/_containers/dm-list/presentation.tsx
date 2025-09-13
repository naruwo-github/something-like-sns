"use client";
import { useState } from "react";
import { getOrCreateDMAction } from "app/_lib/server/actions/dm";
import type { Conversation } from "app/dm/_lib/types";

type Props = {
  initialItems: Conversation[];
};

export default function DMListView({ initialItems }: Props) {
  const [other, setOther] = useState("");

  const onCreate = async () => {
    const otherId = Number(other);
    if (!otherId) return;
    const id = await getOrCreateDMAction(otherId);
    location.href = `/dm/${id}`;
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
        <button type="button" onClick={onCreate} disabled={!other.trim()}>
          作成/開く
        </button>
      </div>
      <ul style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {initialItems.map((c) => (
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
