"use client";
import { useState } from "react";
import { createHeaders, rpc } from "app/_lib/api";
import type { Comment } from "app/post/_lib/types";

type Props = {
  postId: number;
  initialComments: Comment[];
};

export default function PostDetailView({ postId, initialComments }: Props) {
  const [comments, setComments] = useState(initialComments);
  const [body, setBody] = useState("");

  const onSubmit = async () => {
    const r = await rpc<
      { post_id: number; body: string },
      { comment: Comment }
    >(
      "/sns.v1.TimelineService/CreateComment",
      { post_id: postId, body },
      createHeaders(),
    );
    setComments((prev) => [...prev, r.comment]);
    setBody("");
  };

  return (
    <main>
      <h1>Post #{postId}</h1>
      <div style={{ display: "flex", gap: 8, marginBottom: 12 }}>
        <input
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="コメントを書く"
          style={{ flex: 1, padding: 8 }}
        />
        <button type="button" onClick={onSubmit} disabled={!body.trim()}>
          送信
        </button>
      </div>
      <ul style={{ display: "flex", flexDirection: "column", gap: 8 }}>
        {comments.map((c) => (
          <li key={c.id} style={{ border: "1px solid #ddd", padding: 8 }}>
            <div style={{ fontSize: 12, color: "#666" }}>
              by {c.authorUserId} at {new Date(c.createdAt).toLocaleString()}
            </div>
            <div>{c.body}</div>
          </li>
        ))}
      </ul>
    </main>
  );
}
