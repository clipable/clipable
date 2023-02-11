"use client";

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import Link from "next/link";
import { useState, useEffect } from "react";

const inter = Inter({ subsets: ["latin"] });

export default function Home() {
  const loggedIn = false;

  const [videos, setVideos] = useState<Videos[]>([]);

  useEffect(() => {
    const getVids = async () => {
      const vids = await getVideos();
      setVideos(vids);
    };
    getVids();
  }, []);

  console.log(videos);

  return (
    <main className="h-full">
      <div className="container mx-auto py-3">
        <ul role="list" className="grid grid-cols-3 gap-24">
          {videos.map((video) => (
            <li key={video.id}>
              <Link href={`/clips/${video.id}`}>
                <div className="card card-compact w-96 bg-base-100 shadow-xl">
                  <figure>
                    <img src={`http://localhost:8080/api/clips/${video.id}/thumbnail.jpg`} alt="Shoes" />
                  </figure>
                  <div className="card-body">
                    <div className="flex-row flex text-xl">
                      <div className="container flex card-title"> title: {video.title}</div>
                      <div className="container flex card-title"> views: {video.views}</div>
                    </div>
                    <p>{video.description}Desc: {video.description}</p>
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
