"use client";

import { deleteVideo, getVideo, Video } from "@/shared/api";
import { formatDate } from "@/shared/date-formatter";
import { formatViewsCount } from "@/shared/views-formatter";
import dynamic from "next/dynamic";
import { useState, useEffect, useContext } from "react";
import "@mkhuda/react-shaka-player/dist/ui.css";
import Link from "next/link";
import "./player.scss"
import trashcan from "../../../../public/trashcan.svg"
import { UserContext } from "@/context/user-context";
import { useRouter } from "next/navigation";

const ReactShakaPlayer = dynamic(() => import("@mkhuda/react-shaka-player").then(module => module.ReactShakaPlayer), { ssr: false });

export default function Page({ params }: { params: { id: string } }) {
  const [videoDetails, setVideoDetails] = useState<Video | null>(null);
  const userContext = useContext(UserContext);
  const router = useRouter();
  
  useEffect(() => {
    const fetchVideo = async () => {
      const vid = await getVideo(params.id);
      setVideoDetails(vid);
    };
    fetchVideo();
  }, [params.id]);

  return (
    <main className={`mt-2`}>
      <div className="w-fit mx-auto">
        <ReactShakaPlayer
          src={`/api/clips/${params.id}/dash.mpd`}
          autoPlay
        />
      </div>
      {videoDetails && (
        <div className="p-4 mx-auto flex flex-row container">
          <div className="w-full  overflow-hidden text-ellipsis whitespace-nowrap">
            <h1 className="text-2xl font-bold w-[90%] max-w-[90%] overflow-hidden whitespace-nowrap text-ellipsis">{videoDetails.title}</h1>
            <p className="text-gray-300">{videoDetails.description}</p>
          </div>
          <div className="flex-grow"></div>
          <div className="flex flex-row space-x-2 items-center text-gray-400 text-xl">
            <p className="flex flex-row">
              <p className="hover:text-gray-300">
                <Link href={`/users/${videoDetails.creator.id}`}>
                  {videoDetails.creator.username}
                </Link>
              </p>
            </p>
            <p className="text-sm">•</p>
            <p className="whitespace-nowrap">
              {formatViewsCount(videoDetails.views)} view{videoDetails.views === 1 ? "" : "s"}
            </p>
            <p className="text-sm">•</p>
            <p className="whitespace-nowrap">
              {formatDate(videoDetails.created_at)}
            </p>
            {videoDetails.creator.id === userContext.user?.id && (
              <>
                <p className="text-sm">•</p>
                <img src={trashcan.src} className="w-4 text-gray-400 hover:text-gray-300 cursor-pointer" alt="delete clip" title="Delete Clip" onClick={
                  async () => {
                    await deleteVideo(videoDetails.id);
                    router.push("/");
                  }
                } />
              </>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
