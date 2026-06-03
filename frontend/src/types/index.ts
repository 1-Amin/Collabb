export interface User {
  id: string;
  email: string;
  token: string;
}

export interface Board {
  id: string;
  owner_id: string;
  title: string;
  created_at: string;
  columns: Column[];
}

export interface Column {
  id: string;
  board_id: string;
  title: string;
  position: number;
  created_at: string;
  cards: Card[];
}

export interface Card {
  id: string;
  column_id: string;
  title: string;
  description: string;
  position: number;
  created_at: string;
}

export type WsEventType =
  | "card.created"
  | "card.updated"
  | "card.moved"
  | "card.deleted"
  | "column.created"
  | "column.updated"
  | "column.deleted";

export interface WsMessage {
  board_id: string;
  type: WsEventType;
  payload: Card | Column | { id: string };
}
