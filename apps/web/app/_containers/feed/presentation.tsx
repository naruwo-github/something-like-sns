"use client";

import { useState } from "react";
import {
  createPostAction,
  toggleLikeAction,
} from "app/_lib/server/actions/timeline";

type Post = {
  id: number;
  authorUserId: number | string;
  body: string;
  createdAt: string;
  likedByMe: boolean;
  likeCount: number;
  commentCount: number;
};

type Props = { initialPosts: Post[] };

export default function FeedView({ initialPosts }: Props) {
  const [posts, setPosts] = useState<Post[]>(initialPosts);
  const [newBody, setNewBody] = useState("");

  const createPost = async () => {
    const post = await createPostAction(newBody);
    setPosts([post, ...posts]);
    setNewBody("");
  };

  const toggleLike = async (postId: number) => {
    const r = await toggleLikeAction(postId);
    setPosts((prev) =>
      prev.map((p) =>
        p.id === postId ? { ...p, likedByMe: r.active, likeCount: r.total } : p,
      ),
    );
  };

  const moveToPost = (postId: number) => {
    location.href = `/post/${postId}`;
  };

  return (
    <main>
      <h1 style={{ fontSize: 20, fontWeight: "bold", marginBottom: 8 }}>
        Feed (acme)
      </h1>
      <div style={{ display: "flex", gap: 8, marginBottom: 16 }}>
        <input
          value={newBody}
          onChange={(e) => setNewBody(e.target.value)}
          placeholder="いまどうしてる？"
          style={{ flex: 1, padding: 8 }}
        />
        <button type="button" onClick={createPost} disabled={!newBody.trim()}>
          投稿
        </button>
      </div>
      <ul style={{ display: "flex", flexDirection: "column", gap: 12 }}>
        {posts.map((p) => (
          <li
            key={p.id}
            style={{
              border: "1px solid #ddd",
              padding: 12,
              borderRadius: 8,
              cursor: "pointer",
            }}
            onClick={() => moveToPost(Number(p.id))}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                moveToPost(Number(p.id));
              }
            }}
          >
            <div style={{ fontSize: 12, color: "#666" }}>
              by {String(p.authorUserId)} at{" "}
              {new Date(p.createdAt).toLocaleString()}
            </div>
            <div style={{ margin: "8px 0" }}>{p.body}</div>
            <div style={{ display: "flex", gap: 8 }}>
              <button type="button" onClick={() => toggleLike(Number(p.id))}>
                {p.likedByMe ? "♥" : "♡"} {p.likeCount}
              </button>
              <span>💬 {p.commentCount}</span>
            </div>
          </li>
        ))}
      </ul>
    </main>
  );
}
