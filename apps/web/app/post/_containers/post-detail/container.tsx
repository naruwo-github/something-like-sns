"use server";
import { createHeaders, rpc } from "../../_lib/api";
import type { Comment } from "../../_lib/types";
import PostDetailView from "./presentation";

type Props = { postId: number };

export default async function PostDetailContainer({ postId }: Props) {
  await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
  const { items } = await rpc<{ post_id: number }, { items: Comment[] }>(
    "/sns.v1.TimelineService/ListComments",
    { post_id: postId },
    createHeaders(),
  );
  return <PostDetailView postId={postId} initialComments={items} />;
}
