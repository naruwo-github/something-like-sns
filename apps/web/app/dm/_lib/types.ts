export type Conversation = {
  id: number;
  createdAt: string;
  memberUserIds: number[];
};

export type Message = {
  id: number;
  conversationId: number;
  senderUserId: number;
  body: string;
  createdAt: string;
};
