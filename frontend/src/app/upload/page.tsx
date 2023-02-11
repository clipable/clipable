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
  const [title, setTitle] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [file, setFile] = useState<File>();

  const uploadVideo = async () => {
    if (!file || !title || !description) return;

    setState(State.Uploading);

    multiPartForm.append("json", JSON.stringify({ title, description }));
    multiPartForm.append("video", file);

    const response = await fetch("http://localhost:8080/api/clips", {
      method: "POST",
      body: multiPartForm,
      credentials: "include",
    });

    if (response.ok) {
      setState(State.Success);
      setTimeout(() => {
        router.push("/");
      }, 2000);
    } else {
      setState(State.Error);
    }
  };

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
    <main className="h-full">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        {state !== State.Idle && <Alert type={alertType(state)} message={messageBasedOnState(state)} />}
        <div className="form-control w-full max-w-xs">
          <label className="label">
            <span className="label-text">Title</span>
          </label>
          <input
            type="text"
            required
            placeholder="Title"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setTitle(e.target.value);
            }}
          />
          <label className="label">
            <span className="label-text">Description</span>
          </label>
          <input
            type="text"
            required
            placeholder="Description"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setDescription(e.target.value);
            }}
          />
        </div>
        <input
          type="file"
          className="file-input w-full max-w-xs"
          onChange={(e) => {
            if (e.target.files && e.target.files[0]) {
              setFile(e.target.files[0]);
            }
          }}
        />
        <button className="btn btn-primary w-full max-w-xs" onClick={uploadVideo}>
          Upload
        </button>
      </div>
    </main>
  );
}
