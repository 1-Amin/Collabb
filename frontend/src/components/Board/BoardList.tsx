import { useEffect, useState } from "react";
import { api } from "../../api/client";
import type { Board } from "../../types";

interface Props {
  token: string;
  onSelect: (boardID: string) => void;
  onLogout: () => void;
}

export function BoardList({ token, onSelect, onLogout }: Props) {
  const [boards, setBoards] = useState<Board[]>([]);
  const [newTitle, setNewTitle] = useState("");

  useEffect(() => {
    api.listBoards(token).then((b) => setBoards(b ?? []));
  }, [token]);

  async function createBoard() {
    if (!newTitle.trim()) return;
    const b = await api.createBoard(newTitle.trim(), token);
    setBoards((prev) => [...prev, b]);
    setNewTitle("");
  }

  return (
    <div className="board-list-page">
      <header>
        <h1>My Boards</h1>
        <button onClick={onLogout}>Logout</button>
      </header>
      <div className="board-grid">
        {boards.map((b) => (
          <button key={b.id} className="board-card" onClick={() => onSelect(b.id)}>
            {b.title}
          </button>
        ))}
      </div>
      <div className="new-board">
        <input
          placeholder="New board name…"
          value={newTitle}
          onChange={(e) => setNewTitle(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && createBoard()}
        />
        <button onClick={createBoard}>Create Board</button>
      </div>
    </div>
  );
}
