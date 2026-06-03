import { useState } from "react";
import { useDroppable } from "@dnd-kit/core";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { CardItem } from "../Card/CardItem";
import type { Column, Card } from "../../types";

interface Props {
  column: Column;
  onAddCard: (colID: string, title: string, desc: string) => void;
  onUpdateCard: (card: Card, title: string, desc: string) => void;
  onDeleteCard: (card: Card) => void;
  onUpdateTitle: (colID: string, title: string) => void;
  onDelete: (colID: string) => void;
}

export function ColumnItem({
  column,
  onAddCard,
  onUpdateCard,
  onDeleteCard,
  onUpdateTitle,
  onDelete,
}: Props) {
  const [newTitle, setNewTitle] = useState("");
  const [addingCard, setAddingCard] = useState(false);
  const [editingTitle, setEditingTitle] = useState(false);
  const [colTitle, setColTitle] = useState(column.title);

  const { setNodeRef, isOver } = useDroppable({ id: column.id, data: { type: "column" } });

  function submitCard() {
    if (!newTitle.trim()) return;
    onAddCard(column.id, newTitle.trim(), "");
    setNewTitle("");
    setAddingCard(false);
  }

  function saveTitle() {
    onUpdateTitle(column.id, colTitle);
    setEditingTitle(false);
  }

  return (
    <div className={`column ${isOver ? "column-over" : ""}`}>
      <div className="column-header">
        {editingTitle ? (
          <input
            value={colTitle}
            onChange={(e) => setColTitle(e.target.value)}
            onBlur={saveTitle}
            onKeyDown={(e) => e.key === "Enter" && saveTitle()}
            autoFocus
          />
        ) : (
          <h3 onDoubleClick={() => setEditingTitle(true)}>{column.title}</h3>
        )}
        <button className="btn-danger-sm" onClick={() => onDelete(column.id)}>×</button>
      </div>

      <div ref={setNodeRef} className="card-list">
        <SortableContext
          items={column.cards.map((c) => c.id)}
          strategy={verticalListSortingStrategy}
        >
          {column.cards.map((card) => (
            <CardItem
              key={card.id}
              card={card}
              onUpdate={(title, desc) => onUpdateCard(card, title, desc)}
              onDelete={() => onDeleteCard(card)}
            />
          ))}
        </SortableContext>
      </div>

      {addingCard ? (
        <div className="add-card-form">
          <input
            placeholder="Card title"
            value={newTitle}
            onChange={(e) => setNewTitle(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && submitCard()}
            autoFocus
          />
          <div className="card-actions">
            <button onClick={submitCard}>Add</button>
            <button onClick={() => setAddingCard(false)}>Cancel</button>
          </div>
        </div>
      ) : (
        <button className="btn-add-card" onClick={() => setAddingCard(true)}>
          + Add card
        </button>
      )}
    </div>
  );
}
