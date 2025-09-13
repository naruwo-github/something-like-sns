"use server";

import { createHeaders, rpc } from "app/_lib/server/api";

type SendMessageResponse = {
  message: {
    id: number;
    conversationId: number;
    senderUserId: number;
    body: string;
    createdAt: string;
  };
};
type GetOrCreateDMResponse = { conversation_id: number };

export async function sendMessageAction(conversationId: number, body: string) {
  const r = await rpc<
    { conversation_id: number; body: string },
    SendMessageResponse
  >(
    "/sns.v1.DMService/SendMessage",
    { conversation_id: conversationId, body },
    createHeaders(),
  );
  return r.message;
}

export async function getOrCreateDMAction(otherUserId: number) {
  const r = await rpc<{ other_user_id: number }, GetOrCreateDMResponse>(
    "/sns.v1.DMService/GetOrCreateDM",
    { other_user_id: otherUserId },
    createHeaders(),
  );
  return r.conversation_id;
}

