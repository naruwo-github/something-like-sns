import { createHeaders, rpc } from "app/_lib/server/api";
import type { Comment } from "app/post/_lib/types";
import PostDetailView from "./presentation";

type Props = { postId: number };

export default async function PostDetailContainer({ postId }: Props) {
  try {
    await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
    const res = await rpc<
      { post_id: number },
      { items: Comment[] | undefined }
    >(
      "/sns.v1.TimelineService/ListComments",
      { post_id: postId },
      createHeaders(),
    );
    const items = Array.isArray(res?.items) ? res.items : [];
    return <PostDetailView postId={postId} initialComments={items} />;
  } catch (e) {
    console.error(e);
    return <PostDetailView postId={postId} initialComments={[]} />;
  }
}
