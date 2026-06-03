import { useState, useCallback } from "react";
import { api } from "../api/client";
import type { User } from "../types";

const STORAGE_KEY = "collabb_user";

function loadUser(): User | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

export function useAuth() {
  const [user, setUser] = useState<User | null>(loadUser);

  const persist = useCallback((u: User | null) => {
    setUser(u);
    if (u) localStorage.setItem(STORAGE_KEY, JSON.stringify(u));
    else localStorage.removeItem(STORAGE_KEY);
  }, []);

  const login = useCallback(
    async (email: string, password: string) => {
      const res = await api.login(email, password);
      persist({ id: res.id, email: res.email, token: res.token });
    },
    [persist]
  );

  const register = useCallback(
    async (email: string, password: string) => {
      const res = await api.register(email, password);
      persist({ id: res.id, email: res.email, token: res.token });
    },
    [persist]
  );

  const logout = useCallback(() => persist(null), [persist]);

  return { user, login, register, logout };
}
