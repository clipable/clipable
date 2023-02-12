"use client";

import { getVideos, Videos } from "@/shared/api";
import Link from "next/link";
import { useState, useEffect } from "react";

export default function Home() {
  const [videos, setVideos] = useState<Videos[]>([]);

  useEffect(() => {
    const getVids = async () => {
      const vids = await getVideos();
      setVideos(vids);
    };
    getVids();
  }, []);

  return (
    <main className="h-full">
      <div className="container mx-auto py-3">
        <ul role="list" className="grid grid-cols-3 gap-24">
          {videos.map((video) => (
            <li key={video.id}>
              <Link href={`/clips/${video.id}`}>
                <div className="card card-compact w-96 h-full bg-base-100 shadow-xl min-w-[28rem] min-h-[20rem]">
                  <figure className={video.processing ? "flex-auto grow self-center pt-8 glass w-full h-full" : ""}>
                    {video.processing ? (
                      <div>Processing!</div>
                    ) : (
                      <img src={`http://localhost:8080/api/clips/${video.id}/thumbnail.jpg`} />
                    )}
                  </figure>
                  <div className={video.processing ? "card-body justify-end" : "card-body"}>
                    <div className="flex-row flex text-xl">
                      <div className="container flex card-title">{video.title}</div>
                      <div className="container flex card-title justify-end">{video.views}</div>
                    </div>
                    <p className="grow-0">
                      {video.description}
                    </p>
                  </div>
                </div>
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
