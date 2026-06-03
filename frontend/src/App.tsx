import { useState } from "react";
import { useAuth } from "./hooks/useAuth";
import { AuthPage } from "./components/Auth/AuthPage";
import { BoardList } from "./components/Board/BoardList";
import { BoardPage } from "./components/Board/BoardPage";
import "./App.css";

export default function App() {
  const { user, login, register, logout } = useAuth();
  const [activeBoardID, setActiveBoardID] = useState<string | null>(null);

  if (!user) {
    return <AuthPage onLogin={login} onRegister={register} />;
  }

  if (activeBoardID) {
    return (
      <BoardPage
        boardID={activeBoardID}
        token={user.token}
        onBack={() => setActiveBoardID(null)}
      />
    );
  }

  return (
    <BoardList
      token={user.token}
      onSelect={setActiveBoardID}
      onLogout={logout}
    />
  );
}
