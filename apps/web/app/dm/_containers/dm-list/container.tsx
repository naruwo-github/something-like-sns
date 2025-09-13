"use server";
import DMListView from "./presentation";
import { createHeaders, rpc } from "app/_lib/api";
import type { Conversation } from "app/dm/_lib/types";

export default async function DMListContainer() {
  await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
  const { items } = await rpc<Record<string, never>, { items: Conversation[] }>(
    "/sns.v1.DMService/ListConversations",
    {},
    createHeaders(),
  );
  return <DMListView initialItems={items} />;
}
