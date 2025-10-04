import FeedView from "app/_containers/feed/presentation";
import { createHeaders, rpc } from "app/_lib/server/api";

type Post = {
  id: number;
  authorUserId: number | string;
  body: string;
  createdAt: string;
  likedByMe: boolean;
  likeCount: number;
  commentCount: number;
};

export default async function FeedContainer() {
  try {
    await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
    const res = await rpc<Record<string, never>, { items: Post[] | undefined }>(
      "/sns.v1.TimelineService/ListFeed",
      {},
      createHeaders(),
    );
    const items = Array.isArray(res?.items) ? res.items : [];
    return <FeedView initialPosts={items} />;
  } catch (e) {
    console.error(e);
    return <FeedView initialPosts={[]} />;
  }
}
