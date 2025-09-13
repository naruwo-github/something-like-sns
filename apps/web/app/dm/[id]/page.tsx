"use server";
import ConversationContainer from "../_containers/conversation/container";

export default async function Page({ params }: { params: { id: string } }) {
  const cnvId = Number(params.id);
  return <ConversationContainer conversationId={cnvId} />;
}
