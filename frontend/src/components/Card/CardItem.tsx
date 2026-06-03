import { useState } from "react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import type { Card } from "../../types";

interface Props {
  card: Card;
  onUpdate: (title: string, description: string) => void;
  onDelete: () => void;
}

export function CardItem({ card, onUpdate, onDelete }: Props) {
  const [editing, setEditing] = useState(false);
  const [title, setTitle] = useState(card.title);
  const [desc, setDesc] = useState(card.description);

  const { attributes, listeners, setNodeRef, transform, transition, isDragging } =
    useSortable({ id: card.id, data: { type: "card", card } });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  };

  function save() {
    onUpdate(title, desc);
    setEditing(false);
  }

  if (editing) {
    return (
      <div className="card editing" ref={setNodeRef} style={style}>
        <input value={title} onChange={(e) => setTitle(e.target.value)} autoFocus />
        <textarea value={desc} onChange={(e) => setDesc(e.target.value)} rows={3} />
        <div className="card-actions">
          <button onClick={save}>Save</button>
          <button onClick={() => setEditing(false)}>Cancel</button>
        </div>
      </div>
    );
  }

  return (
    <div
      className="card"
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
    >
      <p className="card-title">{card.title}</p>
      {card.description && <p className="card-desc">{card.description}</p>}
      <div className="card-actions">
        <button onClick={(e) => { e.stopPropagation(); setEditing(true); }}>Edit</button>
        <button onClick={(e) => { e.stopPropagation(); onDelete(); }}>Delete</button>
      </div>
    </div>
  );
}
