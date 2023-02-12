"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import Alert from "../../shared/alert";

enum State {
  Idle,
  Error,
  Success,
  Uploading,
  Queued,
  Encoding,
}

export default function Home() {
  const router = useRouter();

  const [state, setState] = useState<State>(State.Idle);
  const [title, setTitle] = useState<string>("");
  const [progress, setProgress] = useState<number>(0)
  const [description, setDescription] = useState<string>("");
  const [file, setFile] = useState<File>();
  const [clipId, setClipId] = useState<string>("");

  const uploadVideo = async () => {
    if (!file || !title) return;

    setState(State.Uploading);

    const multiPartForm = new FormData();

    multiPartForm.append("json", JSON.stringify({ title, description }));
    multiPartForm.append("video", file);

    const req = new XMLHttpRequest();
    req.open("POST", "/api/clips", true);
    req.withCredentials = true;
  
    req.upload.onprogress = (e) => {
      if (e.lengthComputable) {
        setProgress(Math.ceil((e.loaded / e.total) * 100));
      }
    };

    req.onload = () => {
      if (req.readyState === 4) {
        if (req.status === 200) {
          setState(State.Queued);
          setClipId(JSON.parse(req.responseText).id);
        } else {
          setState(State.Error);
        }
      }
    };

    req.send(multiPartForm);
  };

  // A Useeffect hook that loops every 1 second to check progress
  useEffect(() => {
    if (state == State.Queued) {
      setTimeout(checkProgress, 1500);
    }
  }, [clipId]);

  const checkProgress = async () => {
    const rest = await fetch(`/api/clips/progress?cid=${clipId}`);

    if (rest.status === 200) {
      const resp = await rest.json();
      let progress = resp.clips[clipId];

      progress = progress === 0 ? 1 : progress;

      if (progress > 0) {
        setProgress(progress);
      }
      if (state == State.Queued && progress != -1) {
        setState(State.Encoding);
      }
      setTimeout(checkProgress, 1500);
    } else if (rest.status == 204) {
      setState(State.Success);
      // redirect to clip
      router.push(`/clips/${clipId}`);
    }

  };

  const textBasedOnState = (state: State) => {
    switch (state) {
      case State.Error:
        return "Error uploading video";
      case State.Success:
        return "Video uploaded successfully";
      case State.Uploading:
        return "Uploading...";
      case State.Queued:
        return "Queued...";
      case State.Encoding:
        return "Encoding...";
      case State.Idle:
        return "Upload";
      default:
        return "";
    }
  };

  return (
    <main className="h-full">
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        <div className="form-control w-full max-w-xs">
          <label className="label">
            <span className="label-text">Title</span>
          </label>
          <input
            type="text"
            required
            placeholder="Title"
            className="input input-bordered w-full max-w-xs"
            value={title}
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
            value={description}
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
              if (title === "") {
                setTitle(e.target.files[0].name)
              }
            }
          }}
        />
        <button className="btn btn-primary w-full max-w-xs" onClick={uploadVideo}>
          {textBasedOnState(state)}
        </button>
        {state !== State.Idle && <progress className="progress progress-accent w-full max-w-xs" value={progress} max="100"  />}
      </div>
    </main>
  );
}
