import { renderHook, act } from "@testing-library/react";
import { useBoardStore } from "./boardStore";
import type { Board } from "../types";

const makeBoard = (): Board => ({
  id: "b1",
  owner_id: "u1",
  title: "Test Board",
  created_at: "2024-01-01T00:00:00Z",
  columns: [
    {
      id: "col1",
      board_id: "b1",
      title: "Todo",
      position: 0,
      created_at: "2024-01-01T00:00:00Z",
      cards: [
        { id: "card1", column_id: "col1", title: "Task 1", description: "", position: 0, created_at: "" },
        { id: "card2", column_id: "col1", title: "Task 2", description: "", position: 1, created_at: "" },
      ],
    },
    {
      id: "col2",
      board_id: "b1",
      title: "Done",
      position: 1,
      created_at: "2024-01-01T00:00:00Z",
      cards: [],
    },
  ],
});

test("card.created appends to correct column", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  act(() => {
    result.current.applyEvent({
      board_id: "b1",
      type: "card.created",
      payload: { id: "card3", column_id: "col2", title: "New", description: "", position: 0, created_at: "" },
    });
  });
  expect(result.current.board.columns[1].cards).toHaveLength(1);
  expect(result.current.board.columns[1].cards[0].id).toBe("card3");
});

test("card.created is idempotent", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  const payload = { id: "card1", column_id: "col1", title: "Task 1", description: "", position: 0, created_at: "" };
  act(() => {
    result.current.applyEvent({ board_id: "b1", type: "card.created", payload });
    result.current.applyEvent({ board_id: "b1", type: "card.created", payload });
  });
  expect(result.current.board.columns[0].cards).toHaveLength(2);
});

test("card.moved moves across columns", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  act(() => {
    result.current.applyEvent({
      board_id: "b1",
      type: "card.moved",
      payload: { id: "card1", column_id: "col2", title: "Task 1", description: "", position: 0, created_at: "" },
    });
  });
  expect(result.current.board.columns[0].cards).toHaveLength(1);
  expect(result.current.board.columns[1].cards[0].id).toBe("card1");
});

test("card.deleted removes card", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  act(() => {
    result.current.applyEvent({ board_id: "b1", type: "card.deleted", payload: { id: "card1" } });
  });
  expect(result.current.board.columns[0].cards).toHaveLength(1);
  expect(result.current.board.columns[0].cards[0].id).toBe("card2");
});

test("column.created adds column", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  act(() => {
    result.current.applyEvent({
      board_id: "b1",
      type: "column.created",
      payload: { id: "col3", board_id: "b1", title: "In Progress", position: 2, created_at: "", cards: [] },
    });
  });
  expect(result.current.board.columns).toHaveLength(3);
});

test("optimisticMoveCard updates local state immediately", () => {
  const { result } = renderHook(() => useBoardStore(makeBoard()));
  act(() => result.current.optimisticMoveCard("card1", "col2", 0));
  expect(result.current.board.columns[0].cards).toHaveLength(1);
  expect(result.current.board.columns[1].cards[0].id).toBe("card1");
});
