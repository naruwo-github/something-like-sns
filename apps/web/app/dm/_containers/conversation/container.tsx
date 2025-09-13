"use server";
import ConversationView from "./presentation";
import { createHeaders, rpc } from "../../_lib/api";
import type { Message } from "../../_lib/types";

type Props = { conversationId: number };

export default async function ConversationContainer({ conversationId }: Props) {
  await rpc("/sns.v1.TenantService/GetMe", {}, createHeaders());
  const { items } = await rpc<
    { conversation_id: number },
    { items: Message[] }
  >(
    "/sns.v1.DMService/ListMessages",
    { conversation_id: conversationId },
    createHeaders(),
  );
  return (
    <ConversationView conversationId={conversationId} initialItems={items} />
  );
}
