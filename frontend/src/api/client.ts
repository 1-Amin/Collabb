const BASE = import.meta.env.VITE_API_URL ?? "";

async function request<T>(
  path: string,
  options: RequestInit = {},
  token?: string
): Promise<T> {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(options.headers as Record<string, string>),
  };
  if (token) headers["Authorization"] = `Bearer ${token}`;

  const res = await fetch(`${BASE}${path}`, { ...options, headers });
  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || res.statusText);
  }
  if (res.status === 204) return undefined as T;
  return res.json();
}

export const api = {
  // Auth
  register: (email: string, password: string) =>
    request<{ token: string; id: string; email: string }>("/auth/register", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  login: (email: string, password: string) =>
    request<{ token: string; id: string; email: string }>("/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    }),

  // Boards
  listBoards: (token: string) =>
    request<import("../types").Board[]>("/boards", {}, token),

  createBoard: (title: string, token: string) =>
    request<import("../types").Board>("/boards", {
      method: "POST",
      body: JSON.stringify({ title }),
    }, token),

  getBoard: (id: string, token: string) =>
    request<import("../types").Board>(`/boards/${id}`, {}, token),

  // Columns
  createColumn: (boardID: string, title: string, token: string) =>
    request<import("../types").Column>(`/boards/${boardID}/columns`, {
      method: "POST",
      body: JSON.stringify({ title }),
    }, token),

  updateColumn: (colID: string, title: string, token: string) =>
    request<import("../types").Column>(`/columns/${colID}`, {
      method: "PUT",
      body: JSON.stringify({ title }),
    }, token),

  deleteColumn: (colID: string, token: string) =>
    request<void>(`/columns/${colID}`, { method: "DELETE" }, token),

  // Cards
  createCard: (colID: string, boardID: string, title: string, description: string, token: string) =>
    request<import("../types").Card>(
      `/columns/${colID}/cards?board_id=${boardID}`,
      { method: "POST", body: JSON.stringify({ title, description }) },
      token
    ),

  updateCard: (cardID: string, boardID: string, title: string, description: string, token: string) =>
    request<import("../types").Card>(`/cards/${cardID}`, {
      method: "PUT",
      body: JSON.stringify({ title, description, board_id: boardID }),
    }, token),

  moveCard: (cardID: string, boardID: string, columnID: string, position: number, token: string) =>
    request<import("../types").Card>(`/cards/${cardID}/move`, {
      method: "PATCH",
      body: JSON.stringify({ column_id: columnID, position, board_id: boardID }),
    }, token),

  deleteCard: (cardID: string, boardID: string, token: string) =>
    request<void>(`/cards/${cardID}?board_id=${boardID}`, { method: "DELETE" }, token),
};
