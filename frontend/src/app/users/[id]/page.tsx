"use client";

import { getUsersClips, Progress, Clip, ProgressObject } from "@/shared/api";
import VideoCard from "@/shared/clip-card";
import Link from "next/link";
import { useState, useEffect } from "react";

export default function Home({ params }: { params: { id: string } }) {
  const [videos, setVideos] = useState<Clip[]>([]);
  const [videoProgresses, setVideoProgresses] = useState<ProgressObject>({});

  useEffect(() => {
    const getVids = async () => {
      const vids = await getUsersClips(params.id);
      setVideos(vids);
    };
    getVids();
  }, []);

  useEffect(() => {
    const getProgress = async () => {
      if (!videos.length) return;
      const inProgressVideoIds = videos
        .filter((video) => video.processing)
        .map((video) => video.id)
        .join("&cid=");
      if (!inProgressVideoIds) {
        clearInterval(interval);
        return;
      }
      const resp = await fetch(`/api/clips/progress?cid=` + inProgressVideoIds);
      if (resp.status === 204) {
        setVideos(
          videos.map((video) => {
            return { ...video, processing: false };
          })
        );
        clearInterval(interval);
        return;
      } // No content (no videos in progress)
      const { clips } = (await resp.json()) as Progress;
      setVideos(
        videos.map((video) => {
          return { ...video, processing: video.processing && !!clips[video.id] };
        })
      );
      setVideoProgresses(clips);
    };
    const interval = setInterval(getProgress, 2000);
    return () => clearInterval(interval);
  }, [videos]);

  return (
    <main className="h-full">
      <div className="container mx-auto py-3">
        <ul role="list" className="grid grid-cols-3 gap-24">
          {videos.map((video) => (
            <li key={video.id}>
              {video.processing ? (
                <VideoCard video={video} progress={videoProgresses[video.id]} />
              ) : (
                <Link href={`/clips/${video.id}`}>
                  <VideoCard video={video} />
                </Link>
              )}
            </li>
          ))}
        </ul>
      </div>
    </main>
  );
}
