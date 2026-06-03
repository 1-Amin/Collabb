import { useEffect, useRef, useCallback } from "react";
import type { WsMessage } from "../types";

const WS_BASE = import.meta.env.VITE_WS_URL ?? "ws://localhost:8080";

export function useWebSocket(
  boardID: string,
  token: string,
  onMessage: (msg: WsMessage) => void
) {
  const ws = useRef<WebSocket | null>(null);
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>();
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  const connect = useCallback(() => {
    const url = `${WS_BASE}/ws/${boardID}?token=${token}`;
    const socket = new WebSocket(url);
    ws.current = socket;

    socket.onmessage = (e) => {
      // Messages may be batched (newline-separated)
      for (const line of e.data.split("\n")) {
        if (!line.trim()) continue;
        try {
          const msg: WsMessage = JSON.parse(line);
          onMessageRef.current(msg);
        } catch {
          // ignore malformed
        }
      }
    };

    socket.onclose = () => {
      // Reconnect with exponential back-off (capped at 10s)
      reconnectTimer.current = setTimeout(connect, Math.min(10000, 1000));
    };

    socket.onerror = () => socket.close();
  }, [boardID, token]);

  useEffect(() => {
    connect();
    return () => {
      clearTimeout(reconnectTimer.current);
      ws.current?.close();
    };
  }, [connect]);
}
