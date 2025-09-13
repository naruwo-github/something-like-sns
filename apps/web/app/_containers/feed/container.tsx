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
  await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
  const { items } = await rpc<Record<string, never>, { items: Post[] }>(
    "/sns.v1.TimelineService/ListFeed",
    {},
    createHeaders(),
  );
  return <FeedView initialPosts={items} />;
}
