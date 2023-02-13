"use client";

import { getVideos, Video } from "@/shared/api";
import VideoCard from "@/shared/video-card";
import Link from "next/link";
import { useState, useEffect } from "react";

export default function Home() {
  const [videos, setVideos] = useState<Video[]>([]);

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
                <VideoCard video={video} />
              </Link>
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
