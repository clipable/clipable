"use client";

import { FormEvent, useContext, useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { register, registrationAllowed } from "@/shared/api";
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
  const [isRegistrationAllowed, setRegistrationAllowed] = useState<boolean | null>(null);

  useEffect(() => {
    const getRegistrationAllowed = async () => {
      setRegistrationAllowed(await registrationAllowed());
    };
    getRegistrationAllowed();
  }, []);

  const registerUser = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

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

  const messageBasedOnState = (state: State) => {
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

  if (isRegistrationAllowed != null && !isRegistrationAllowed) {
    return (
      <main className="h-screen">
        <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
          <Alert type="error" message="Registration is disabled for this site" />
        </div>
      </main>
    );
  }

  return (
    <main className="h-screen">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        {state !== State.Idle && (
          <Alert type={state === State.Success ? "success" : "error"} message={messageBasedOnState(state)} />
        )}
        <form className="form-control w-full max-w-xs" onSubmit={registerUser} id="registerForm">
          <label className="label" htmlFor="username">
            <span className="label-text">Username</span>
          </label>
          <input
            type="text"
            placeholder="Username"
            id="username"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setUsername(e.target.value);
            }}
          />
          <label className="label" htmlFor="password">
            <span className="label-text">Password</span>
          </label>
          <input
            type="password"
            placeholder="Password"
            id="password"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setPassword(e.target.value);
            }}
          />
        </form>
        <button className="btn btn-primary w-full max-w-xs" form="registerForm">
          Register
        </button>
      </div>
    </main>
  );
}
