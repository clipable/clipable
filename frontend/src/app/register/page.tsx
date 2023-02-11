"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Alert from "../../shared/alert";

enum State {
  Idle,
  Error,
  Success,
  Uploading,
}

export default function Home() {
  const multiPartForm = new FormData();

  const router = useRouter();
  
  const [state, setState] = useState<State>(State.Idle);
  const [username, setUsername] = useState<string>("");
  const [password, setPassword] = useState<string>("");

  const uploadVideo = async () => {};

  const messageBasedOnState = (state: State) => {
    switch (state) {
      case State.Error:
        return "Error uploading video";
      case State.Success:
        return "Video uploaded successfully";
      case State.Uploading:
        return "Uploading video...";
      default:
        return "";
    }
  };

  const alertType = (state: State) => {
    switch (state) {
      case State.Error:
        return "error";
      case State.Success:
        return "success";
      case State.Uploading:
        return "info";
      default:
        return "info";
    }
  };

  return (
    <main className="h-screen">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        {state !== State.Idle && <Alert type={alertType(state)} message={messageBasedOnState(state)} />}
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
            type="text"
            placeholder="Password"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setPassword(e.target.value);
            }}
          />
        </div>
        <button
          className="btn btn-primary w-full max-w-xs"
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
