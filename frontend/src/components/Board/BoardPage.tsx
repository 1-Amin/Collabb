import { useEffect, useState, useCallback } from "react";
import {
  DndContext,
  DragEndEvent,
  DragOverEvent,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
} from "@dnd-kit/core";
import { api } from "../../api/client";
import { useWebSocket } from "../../hooks/useWebSocket";
import { useBoardStore } from "../../store/boardStore";
import { ColumnItem } from "../Column/ColumnItem";
import type { Board, Card, WsMessage } from "../../types";

interface Props {
  boardID: string;
  token: string;
  onBack: () => void;
}

export function BoardPage({ boardID, token, onBack }: Props) {
  const [loading, setLoading] = useState(true);
  const [initialBoard, setInitialBoard] = useState<Board | null>(null);

  useEffect(() => {
    api.getBoard(boardID, token).then((b) => {
      setInitialBoard(b);
      setLoading(false);
    });
  }, [boardID, token]);

  if (loading || !initialBoard) return <div className="loading">Loading…</div>;
  return <BoardView board={initialBoard} token={token} onBack={onBack} />;
}

function BoardView({ board: initial, token, onBack }: { board: Board; token: string; onBack: () => void }) {
  const { board, applyEvent, optimisticAddCard, optimisticMoveCard, optimisticAddColumn } =
    useBoardStore(initial);
  const [newColTitle, setNewColTitle] = useState("");

  const handleWsMessage = useCallback(
    (msg: WsMessage) => applyEvent(msg),
    [applyEvent]
  );
  useWebSocket(board.id, token, handleWsMessage);

  const sensors = useSensors(useSensor(PointerSensor, { activationConstraint: { distance: 5 } }));

  // ── Columns ──────────────────────────────────────────────────────────────

  async function addColumn() {
    if (!newColTitle.trim()) return;
    const col = await api.createColumn(board.id, newColTitle.trim(), token);
    optimisticAddColumn(col); // server also broadcasts; idempotent guard in store
    setNewColTitle("");
  }

  async function updateColumnTitle(colID: string, title: string) {
    await api.updateColumn(colID, title, token);
    // broadcast arrives via WS
  }

  async function deleteColumn(colID: string) {
    await api.deleteColumn(colID, token);
    applyEvent({ board_id: board.id, type: "column.deleted", payload: { id: colID } });
  }

  // ── Cards ─────────────────────────────────────────────────────────────────

  async function addCard(colID: string, title: string, desc: string) {
    const tempCard = {
      id: `temp-${Date.now()}`,
      column_id: colID,
      title,
      description: desc,
      position: 999,
      created_at: new Date().toISOString(),
    };
    optimisticAddCard(tempCard);
    const card = await api.createCard(colID, board.id, title, desc, token);
    // Replace temp with real
    applyEvent({ board_id: board.id, type: "card.created", payload: card });
    applyEvent({ board_id: board.id, type: "card.deleted", payload: { id: tempCard.id } });
  }

  async function updateCard(card: Card, title: string, desc: string) {
    await api.updateCard(card.id, board.id, title, desc, token);
  }

  async function deleteCard(card: Card) {
    applyEvent({ board_id: board.id, type: "card.deleted", payload: { id: card.id } });
    await api.deleteCard(card.id, board.id, token);
  }

  // ── Drag & Drop ───────────────────────────────────────────────────────────

  function findCardAndColumn(cardID: string) {
    for (const col of board.columns) {
      const card = col.cards.find((c) => c.id === cardID);
      if (card) return { card, col };
    }
    return null;
  }

  function onDragOver(e: DragOverEvent) {
    const { active, over } = e;
    if (!over || active.id === over.id) return;

    const activeData = active.data.current as { type: string; card?: Card };
    if (activeData?.type !== "card") return;

    const overID = String(over.id);
    // Dragging over a column droppable
    const overCol = board.columns.find((c) => c.id === overID);
    if (overCol) {
      optimisticMoveCard(String(active.id), overCol.id, overCol.cards.length);
      return;
    }
    // Dragging over another card
    const found = findCardAndColumn(overID);
    if (found) {
      optimisticMoveCard(String(active.id), found.col.id, found.card.position);
    }
  }

  async function onDragEnd(e: DragEndEvent) {
    const { active, over } = e;
    if (!over) return;

    const activeData = active.data.current as { type: string; card?: Card };
    if (activeData?.type !== "card") return;

    const cardID = String(active.id);
    const overID = String(over.id);

    // Find destination column and position
    let destColID: string;
    let destPos: number;

    const overCol = board.columns.find((c) => c.id === overID);
    if (overCol) {
      destColID = overCol.id;
      destPos = overCol.cards.length;
    } else {
      const found = findCardAndColumn(overID);
      if (!found) return;
      destColID = found.col.id;
      destPos = found.card.position;
    }

    const src = findCardAndColumn(cardID);
    if (src && src.col.id === destColID && src.card.position === destPos) return;

    await api.moveCard(cardID, board.id, destColID, destPos, token);
  }

  return (
    <div className="board-page">
      <header className="board-header">
        <button onClick={onBack}>← Boards</button>
        <h2>{board.title}</h2>
      </header>

      <DndContext
        sensors={sensors}
        collisionDetection={closestCorners}
        onDragOver={onDragOver}
        onDragEnd={onDragEnd}
      >
        <div className="board-columns">
          {board.columns.map((col) => (
            <ColumnItem
              key={col.id}
              column={col}
              onAddCard={addCard}
              onUpdateCard={updateCard}
              onDeleteCard={deleteCard}
              onUpdateTitle={updateColumnTitle}
              onDelete={deleteColumn}
            />
          ))}

          <div className="add-column">
            <input
              placeholder="New column…"
              value={newColTitle}
              onChange={(e) => setNewColTitle(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && addColumn()}
            />
            <button onClick={addColumn}>Add column</button>
          </div>
        </div>
      </DndContext>
    </div>
  );
}
