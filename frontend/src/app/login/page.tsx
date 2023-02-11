"use client";

import { useContext, useState } from "react";
import { useRouter } from "next/navigation";
import { login } from "@/shared/api";
import { UserContext } from "@/context/user-context";

export default function Home() {
  const router = useRouter();
  const userContext = useContext(UserContext);

  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const loginUser = async () => {
    const ok = await login(username, password);
    if (ok) {
      userContext.reload();
      router.push("/");
    }
  };

  return (
    <main className="h-screen">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
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
        <button className="btn btn-primary w-full max-w-xs" onClick={loginUser}>
          Login
        </button>

        <button
          className="btn btn-outline w-full max-w-xs"
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
