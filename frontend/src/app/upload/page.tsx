"use client";

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import Link from "next/link";
import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";

const inter = Inter({ subsets: ["latin"] });

export default function Home() {
  const multiPartForm = new FormData();

  const router = useRouter();

  const [title, setTitle] = useState<string>("");
  const [description, setDescription] = useState<string>("");
  const [file, setFile] = useState<File>();

  const uploadVideo = async () => {
    console.log("Uploading video...");
    console.log(title, description, file);
    if (!file || !title || !description) return;

    multiPartForm.append("json", JSON.stringify({ title, description }));
    multiPartForm.append("video", file);

    const response = await fetch("http://localhost:8080/api/clips", {
      method: "POST",
      body: multiPartForm,
      credentials: "include",
    });

    if (response.ok) {
      console.log("Uploaded!");
      router.push("/");
    }
  };

  return (
    <main className="h-screen">
      <header className="navbar bg-base-300">
        <nav className="flex px-2 lg:px-8" aria-label="Top">
          <div className="flex w-full grow justify-between border-b border-indigo-500 py-1 lg:border-none">
            <div className="flex items-center">
              <Link href="/">
                <span className="dark:text-white font-bold">Clipable</span>
              </Link>
            </div>
          </div>
        </nav>
      </header>
      <div className="container mx-auto flex flex-col space-y-6 justify-center items-center py-3">
        <div className="form-control w-full max-w-xs">
          <label className="label">
            <span className="label-text">Title</span>
          </label>
          <input
            type="text"
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
            placeholder="Description"
            className="input input-bordered w-full max-w-xs"
            onChange={(e) => {
              setDescription(e.target.value);
            }}
          />
        </div>
        <input type="file" className="file-input w-full max-w-xs"
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
