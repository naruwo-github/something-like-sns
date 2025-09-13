import ConversationView from "app/dm/_containers/conversation/presentation";
import { createHeaders, rpc } from "app/_lib/server/api";
import type { Message } from "app/dm/_lib/types";

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
    <ConversationView
      conversationId={conversationId}
      initialItems={items.sort((a, b) => a.id - b.id)}
    />
  );
}
