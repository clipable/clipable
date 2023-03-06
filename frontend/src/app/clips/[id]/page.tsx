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
      <div className="w-fit mx-auto ">
      <ReactShakaPlayer onLoad={(player) => setMainPlayer(player)} autoPlay />
        {videoDetails && (
          <div className="w-full bg-base-300 mx-0 flex-grow flex flex-row">
            <div className="w-2/4 flex-grow overflow-hidden text-ellipsis whitespace-nowrap">
              <h1 className="indent-1.5 text-[15px] font-semibold overflow-hidden whitespace-nowrap text-ellipsis">
                {videoDetails.title}
              </h1>
              <p className="whitespace-pre-line truncate indent-1.5 text-[10px] text-gray-300">
                {videoDetails.unlisted && <div className="badge badge-outline">unlisted</div>}
                {" " + videoDetails.description?.slice(0, 40)+ "..."}
                </p>
            </div>
            <div className=""></div>
            <div className="w-1/4 flex-col items-center text-gray-400 pr-3">
              <p className="flex">
                <p className="flex-grow"></p>
                <p className="hover:text-gray-300 text-[12px]">
                  <Link href={`/users/${videoDetails.creator.id}`}>{videoDetails.creator.username}</Link>
                </p>
              </p>

              <p className="whitespace-nowrap text-[8px] text-right">
                {formatViewsCount(videoDetails.views)} view{videoDetails.views === 1 ? "" : "s - "}
                {formatDate(videoDetails.created_at)}
              </p>
            </div>
            {videoDetails.creator.id === userContext.user?.id && (
              <>
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
        )}
      </div>
    </main>
  );
}
