import { useState, useCallback } from "react";
import type { Board, Card, Column, WsMessage } from "../types";

export function useBoardStore(initial: Board) {
  const [board, setBoard] = useState<Board>(initial);

  const applyEvent = useCallback((msg: WsMessage) => {
    setBoard((prev) => {
      switch (msg.type) {
        case "column.created": {
          const col = msg.payload as Column;
          if (prev.columns.some((c) => c.id === col.id)) return prev;
          return { ...prev, columns: [...prev.columns, { ...col, cards: [] }] };
        }
        case "column.updated": {
          const col = msg.payload as Column;
          return {
            ...prev,
            columns: prev.columns.map((c) =>
              c.id === col.id ? { ...c, title: col.title } : c
            ),
          };
        }
        case "column.deleted": {
          const { id } = msg.payload as { id: string };
          return { ...prev, columns: prev.columns.filter((c) => c.id !== id) };
        }
        case "card.created": {
          const card = msg.payload as Card;
          return {
            ...prev,
            columns: prev.columns.map((col) =>
              col.id === card.column_id
                ? {
                    ...col,
                    cards: col.cards.some((c) => c.id === card.id)
                      ? col.cards
                      : [...col.cards, card],
                  }
                : col
            ),
          };
        }
        case "card.updated": {
          const card = msg.payload as Card;
          return {
            ...prev,
            columns: prev.columns.map((col) => ({
              ...col,
              cards: col.cards.map((c) => (c.id === card.id ? card : c)),
            })),
          };
        }
        case "card.moved": {
          const card = msg.payload as Card;
          return {
            ...prev,
            columns: prev.columns.map((col) => {
              // Remove from old column
              const filtered = col.cards.filter((c) => c.id !== card.id);
              if (col.id === card.column_id) {
                // Insert at new position
                const next = [...filtered];
                next.splice(card.position, 0, card);
                return { ...col, cards: next };
              }
              return { ...col, cards: filtered };
            }),
          };
        }
        case "card.deleted": {
          const { id } = msg.payload as { id: string };
          return {
            ...prev,
            columns: prev.columns.map((col) => ({
              ...col,
              cards: col.cards.filter((c) => c.id !== id),
            })),
          };
        }
        default:
          return prev;
      }
    });
  }, []);

  // Optimistic helpers — mutate local state immediately, API confirms
  const optimisticAddCard = useCallback((card: Card) => {
    setBoard((prev) => ({
      ...prev,
      columns: prev.columns.map((col) =>
        col.id === card.column_id && !col.cards.some((c) => c.id === card.id)
          ? { ...col, cards: [...col.cards, card] }
          : col
      ),
    }));
  }, []);

  const optimisticMoveCard = useCallback(
    (cardID: string, toColID: string, toPos: number) => {
      setBoard((prev) => {
        let moved: Card | undefined;
        const withoutCard = prev.columns.map((col) => {
          const card = col.cards.find((c) => c.id === cardID);
          if (card) moved = { ...card, column_id: toColID, position: toPos };
          return { ...col, cards: col.cards.filter((c) => c.id !== cardID) };
        });
        if (!moved) return prev;
        return {
          ...prev,
          columns: withoutCard.map((col) => {
            if (col.id !== toColID) return col;
            const next = [...col.cards];
            next.splice(toPos, 0, moved!);
            return { ...col, cards: next };
          }),
        };
      });
    },
    []
  );

  const optimisticAddColumn = useCallback((col: Column) => {
    setBoard((prev) => {
      if (prev.columns.some((c) => c.id === col.id)) return prev;
      return { ...prev, columns: [...prev.columns, { ...col, cards: [] }] };
    });
  }, []);

  return { board, setBoard, applyEvent, optimisticAddCard, optimisticMoveCard, optimisticAddColumn };
}
