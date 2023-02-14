"use client";

import { getVideos, Progress, Video } from "@/shared/api";
import VideoCard from "@/shared/video-card";
import Link from "next/link";
import { useState, useEffect } from "react";

export default function Home() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [videoProgresses, setVideoProgresses] = useState<Record<string, number>>({})

  useEffect(() => {
    const getVids = async () => {
      const vids = await getVideos();
      setVideos(vids);
    };
    getVids();
  }, []);

  useEffect(() => {
    const getProgress = async () => {
      if (!videos.length) return

      const inProgressVideoIds = videos.filter(video => video.processing).map(video => video.id).join('&cid=')
      const resp = await fetch(`/api/clips/progress?cid=` + inProgressVideoIds);
      const { clips } = (await resp.json()) as Progress
      setVideoProgresses(clips)
    }
    const interval = setInterval(getProgress, 1000);
    return () => clearInterval(interval);
  }, [videos]);

  console.log(videoProgresses)

  return (
    <main className="h-full">
      <div className="container mx-auto py-3">
        <ul role="list" className="grid grid-cols-3 gap-24">
          {videos.map((video) => (
            <li key={video.id}>
              {video.processing
                ? <VideoCard video={video} progress={videoProgresses[video.id]}  />
                : <Link href={`/clips/${video.id}`}>
                  <VideoCard video={video} />
                </Link>
              }
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
