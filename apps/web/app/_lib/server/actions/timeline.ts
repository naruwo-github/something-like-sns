"use server";

import { rpc, createHeaders } from "app/_lib/server/api";

type CreatePostResponse = {
  post: {
    id: number;
    likedByMe: boolean;
    likeCount: number;
    commentCount: number;
    authorUserId: number | string;
    body: string;
    createdAt: string;
  };
};
type ToggleReactionResponse = { active: boolean; total: number };

export async function createPostAction(body: string) {
  const r = await rpc<{ body: string }, CreatePostResponse>(
    "/sns.v1.TimelineService/CreatePost",
    { body },
    createHeaders(),
  );
  return r.post;
}

export async function toggleLikeAction(postId: number) {
  const r = await rpc<
    { targetType: number; targetId: number; type: string },
    ToggleReactionResponse
  >(
    "/sns.v1.ReactionService/ToggleReaction",
    { targetType: 1, targetId: postId, type: "like" },
    createHeaders(),
  );
  return r;
}
