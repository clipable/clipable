"use client";

import { FormEvent, useContext, useState } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/shared/api";
import { UserContext } from "@/context/user-context";

export default function Home() {
  const router = useRouter();
  const userContext = useContext(UserContext);

  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const loginUser = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const ok = await login(username, password);
    if (ok) {
      userContext.reload();
      router.push("/");
    }
  };

  return (
    <main className="h-screen">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        <form className="form-control w-full max-w-xs" onSubmit={loginUser} id="loginForm">
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

        <button className="btn btn-primary w-full max-w-xs" form="loginForm">
          Login
        </button>

        <button
          className="btn btn-outline w-full max-w-xs"
          type="button"
          onClick={() => {
            router.push("/register");
          }}
        >
          Register
        </button>
      </div>
    </main>
  );
}
