"use client";

import { deleteCip, getClip, Clip } from "@/shared/api";
import { formatDate } from "@/shared/date-formatter";
import { formatViewsCount } from "@/shared/views-formatter";
import dynamic from "next/dynamic";
import { useState, useEffect, useContext } from "react";
import "@mkhuda/react-shaka-player/dist/ui.css";
import Link from "next/link";
import "./player.scss";
import trashcan from "../../../../public/trashcan.svg";
import { UserContext } from "@/context/user-context";
import { useRouter } from "next/navigation";

const ReactShakaPlayer = dynamic(() => import("@mkhuda/react-shaka-player").then((module) => module.ReactShakaPlayer), {
  ssr: false,
});

export default function Page({ params }: { params: { id: string } }) {
  const [videoDetails, setVideoDetails] = useState<Clip | null>(null);
  const userContext = useContext(UserContext);
  const router = useRouter();
  let [mainPlayer, setMainPlayer] = useState({});

  useEffect(() => {
    const { player, videoElement } = mainPlayer as { player: any; videoElement: HTMLVideoElement };
    if (!player) return;

    const play = async () => {
      videoElement.onvolumechange = () => {
        localStorage.setItem("volume", videoElement.volume.toString());
      };
      await player.load(`/api/clips/${params.id}/dash.mpd`);
      videoElement.volume = parseFloat(localStorage.getItem("volume") || "1");
      videoElement.play();
    };
    play();
    return () => {
      player.unload();
      videoElement.onvolumechange = null;
    };
  }, [mainPlayer]);

  useEffect(() => {
    const fetchVideo = async () => {
      const vid = await getClip(params.id);
      setVideoDetails(vid);
    };
    fetchVideo();
  }, [params.id]);

  return (
    <main className={`mt-2`}>
      <div className="w-fit mx-auto">
        <ReactShakaPlayer onLoad={(player) => setMainPlayer(player)} autoPlay />
      </div>
      {videoDetails && (
        <div className="p-4 mx-auto flex flex-row container">
          <div className="w-full  overflow-hidden text-ellipsis whitespace-nowrap">
            <div className="flex flex-row space-x-4 items-center">
              <h1 className="text-2xl font-bold max-w-[90%] overflow-hidden whitespace-nowrap text-ellipsis">
                {videoDetails.title}
              </h1>
            </div>
            {videoDetails.unlisted && <div className="badge badge-outline">unlisted</div>}
            <p className="text-gray-300">{videoDetails.description}</p>
          </div>
          <div className="flex-grow"></div>
          <div className="flex flex-row self-start space-x-2 items-center text-gray-400 text-xl">
            <p className="flex flex-row">
              <p className="hover:text-gray-300">
                <Link href={`/users/${videoDetails.creator.id}`}>{videoDetails.creator.username}</Link>
              </p>
            </p>
            <p className="text-sm">•</p>
            <p className="whitespace-nowrap">
              {formatViewsCount(videoDetails.views)} view{videoDetails.views === 1 ? "" : "s"}
            </p>
            <p className="text-sm">•</p>
            <p className="whitespace-nowrap">{formatDate(videoDetails.created_at)}</p>
            {videoDetails.creator.id === userContext.user?.id && (
              <>
                <p className="text-sm">•</p>
                <img
                  src={trashcan.src}
                  className="w-4 cursor-pointer"
                  alt="delete clip"
                  title="Delete Clip"
                  onClick={async () => {
                    await deleteCip(videoDetails.id);
                    router.push("/");
                  }}
                />
              </>
            )}
          </div>
        </div>
      )}
    </main>
  );
}
