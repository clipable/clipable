"use client";

import { useContext, useState } from "react";
import { useRouter } from "next/navigation";
import { register } from "@/shared/api";
import { UserContext } from "@/context/user-context";
import Alert from "@/shared/alert";

enum State {
  Idle,
  UserExists,
  UnknownError,
  Success,
}

export default function Home() {
  const router = useRouter();
  const userContext = useContext(UserContext);

  const [state, setState] = useState<State>(State.Idle);
  const [username, setUsername] = useState<string | null>(null);
  const [password, setPassword] = useState<string | null>(null);

  const registerUser = async () => {
    if (!username || !password) return;
    const resp = await register(username, password);
    if (resp.ok) {
      setState(State.Success);
      setTimeout(() => {
        userContext.reload();
        router.push("/");
      }, 2000);
      return;
    }
    if (resp.status === 409) {
      setState(State.UserExists);
      return;
    }

    setState(State.UnknownError);
  };

  const messageFromState = (state: State) => {
    switch (state) {
      case State.UserExists:
        return "Someone with that name already exists";
      case State.UnknownError:
        return "Something went wrong";
      case State.Success:
        return "Successfully registered, redirecting...";
      default:
        return "";
    }
  };

  return (
    <main className="h-screen">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        {state !== State.Idle && (
          <Alert type={state === State.Success ? "success" : "error"} message={messageFromState(state)} />
        )}
        <div className="form-control w-full max-w-xs">
          <label className="label">
            <span className="label-text">Username</span>
          </label>
          <input
            type="text"
            placeholder="Username"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setUsername(e.target.value);
            }}
          />
          <label className="label">
            <span className="label-text">Password</span>
          </label>
          <input
            type="password"
            placeholder="Password"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setPassword(e.target.value);
            }}
          />
        </div>
        <button className="btn btn-primary w-full max-w-xs" onClick={registerUser}>
          Register
        </button>
      </div>
    </main>
  );
}
