"use client";

import { getVideo, Video } from "@/shared/api";
import { formatDate } from "@/shared/date-formatter";
import { formatViewsCount } from "@/shared/views-formatter";
import dynamic from "next/dynamic";
import { useState, useEffect } from "react";

const ShakaPlayer = dynamic(() => import("shaka-player-react"), { ssr: false });

import "shaka-player-react/dist/controls.css";

export default function Page({ params }: { params: { id: string } }) {
  const [videoDetails, setVideoDetails] = useState<Video | null>(null);

  useEffect(() => {
    const fetchVideo = async () => {
      const vid = await getVideo(params.id);
      setVideoDetails(vid);
    };
    fetchVideo();
  }, [params.id]);

  return (
    <main className="max-w-[70%] mt-6 mx-auto">
      <ShakaPlayer src={`/api/clips/${params.id}/dash.mpd`} autoPlay />
      {videoDetails && (
        <div className="p-4 flex flex-row">
          <div>
            <h1 className="text-2xl font-bold">{videoDetails.title}</h1>
            <p className="text-gray-300">{videoDetails.description}</p>
          </div>
          <div className="flex-grow"></div>
          <div className="flex flex-row space-x-2 text-gray-400 text-xl">
            <p>{formatDate(videoDetails.created_at)}</p>
            <p>•</p>
            <p>
              {formatViewsCount(videoDetails.views)} view{videoDetails.views === 1 ? "" : "s"}
            </p>
          </div>
        </div>
      )}
    </main>
  );
}
