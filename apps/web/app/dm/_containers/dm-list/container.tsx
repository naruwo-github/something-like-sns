import DMListView from "app/dm/_containers/dm-list/presentation";
import { createHeaders, rpc } from "app/_lib/server/api";
import type { Conversation } from "app/dm/_lib/types";

export default async function DMListContainer() {
  try {
    await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
    const res = await rpc<
      Record<string, never>,
      { items: Conversation[] | undefined }
    >("/sns.v1.DMService/ListConversations", {}, createHeaders());
    const items = Array.isArray(res?.items) ? res.items : [];
    return <DMListView initialItems={items} />;
  } catch (e) {
    console.error(e);
    return <DMListView initialItems={[]} />;
  }
}
