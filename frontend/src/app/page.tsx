"use client";

import { getVideos, Progress, Video, ProgressObject, searchVideos } from "@/shared/api";
import VideoCard from "@/shared/video-card";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState, useEffect } from "react";

export default function Home() {
  const [videos, setVideos] = useState<Video[] | null>(null);
  const [videoProgresses, setVideoProgresses] = useState<ProgressObject>({});

  const params = useSearchParams();

  useEffect(() => {
    const getVids = async () => {
      const vids = await getVideos();
      setVideos(vids);
    };
    const getSearchedVids = async () => {
      const vids = await searchVideos(params.get("search") as string);
      setVideos(vids);
    };
    params.get("search") ? getSearchedVids() : getVids();
  }, [params]);

  useEffect(() => {
    if (!videos) return;
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
      <div className="flex justify-center py-3 mx-3">
        {videos?.length === 0 && (
          <div className="flex flex-col items-center justify-center w-full">
            <h1 className="text-4xl font-bold">No videos found</h1>
            {params.get("search") && <p className="text-xl">Try searching for something else</p>}
          </div>
        )}
        {videos && videos.length > 0 && (
          <ul role="list" className="grid grid-cols-1 xl:grid-cols-2 2xl:grid-cols-3 gap-10 xl:gap-16 2xl:gap-24">
            {videos.map((video) => (
              <li className="m-0" key={video.id}>
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
        )}
      </div>
    </main>
  );
}
