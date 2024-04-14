"use client";

import { getClips, Progress, Clip, ProgressObject, searchClips } from "@/shared/api";
import ClipCard from "@/shared/clip-card";
import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useState, useEffect, useRef } from "react";
import Footer from "./footer";

export default function Home() {
  const [videos, setVideos] = useState<Clip[] | null>(null);
  const [videoProgresses, setVideoProgresses] = useState<ProgressObject>({});

  const params = useSearchParams();

  useEffect(() => {
    const getVids = async () => {
      const vids = await getClips();
      setVideos(vids);
    };
    const getSearchedVids = async () => {
      const vids = await searchClips(params.get("search") as string);
      setVideos(vids);
    };
    params.get("search") ? getSearchedVids() : getVids();
  }, [params]);

  useEffect(() => {
    if (!videos) return;
    const getProgress = async () => {
      // If there are no videos stop the interval
      if (!videos.length) {
        clearInterval(interval);
        return;
      }

      // Get the ids of all videos that are still processing
      const inProgressVideoIds = videos
        .filter((video) => video.processing)
        .map((video) => video.id)
        .join("&cid=");

      // If there are no videos in progress stop the interval
      if (!inProgressVideoIds) {
        clearInterval(interval);
        return;
      }

      const resp = await fetch(`/api/clips/progress?cid=` + inProgressVideoIds);

      // If the request fails, we don't want to stop the interval
      if (!resp.ok) {
        return;
      }

      // If the server returns 204, there are no videos in progress anymore
      // so we can stop the interval and set all videos to not processing
      if (resp.status === 204) {
        setVideos(
          videos.map((video) => {
            return { ...video, processing: false };
          })
        );
        clearInterval(interval);
        return;
      }

      const { clips } = (await resp.json()) as Progress;
      setVideos(
        videos
          // If the clip progress is -2, it failed to encode, so we don't want to show it
          .filter((video) => clips[video.id] !== -2)
          // If the clip is still processing, but we don't have a progress value for it, set it to be done processing
          .map((video) => ({ ...video, processing: video.processing && !!clips[video.id] }))
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
                  <ClipCard video={video} progress={videoProgresses[video.id]} />
                ) : (
                  <Link href={`/clips/${video.id}`}>
                    <ClipCard video={video} />
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
