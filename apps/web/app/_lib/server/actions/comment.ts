"use server";

import { createHeaders, rpc } from "app/_lib/server/api";

type CreateCommentResponse = {
  comment: {
    id: number;
    postId: number;
    authorUserId: number;
    body: string;
    createdAt: string;
  };
};

export async function createCommentAction(postId: number, body: string) {
  const r = await rpc<{ post_id: number; body: string }, CreateCommentResponse>(
    "/sns.v1.TimelineService/CreateComment",
    { post_id: postId, body },
    createHeaders(),
  );
  return r.comment;
}

