"use client";

import { deleteCip, getClip, Clip, updateClipDetails } from "@/shared/api";
import { formatDate } from "@/shared/date-formatter";
import { formatViewsCount } from "@/shared/views-formatter";
import dynamic from "next/dynamic";
import { useState, useEffect, useContext } from "react";
import "@mkhuda/react-shaka-player/dist/ui.css";
import Link from "next/link";
import "./player.scss";
import { UserContext } from "@/context/user-context";
import { useRouter } from "next/navigation";
import Modal from "@/shared/modal";
import { PencilSquareIcon, TrashIcon } from "@heroicons/react/24/outline";

const ReactShakaPlayer = dynamic(() => import("@mkhuda/react-shaka-player").then((module) => module.ReactShakaPlayer), {
  ssr: false,
});

export default function Page({ params }: { params: { id: string } }) {
  const [videoDetails, setVideoDetails] = useState<Clip | null>(null);
  const userContext = useContext(UserContext);
  const router = useRouter();
  let [mainPlayer, setMainPlayer] = useState({});
  const [deleteModalOpen, setDeleteModalOpen] = useState(false);
  const [editModalOpen, setEditModalOpen] = useState(false);
  const [videoTitle, setVideoTitle] = useState("");
  const [videoDescription, setVideoDescription] = useState("");
  const [videoUnlisted, setVideoUnlisted] = useState(false);

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

  const fetchVideo = async () => {
    const vid = await getClip(params.id);
    if (!vid) return;
    setVideoDetails(vid);
    setVideoTitle(vid.title);
    setVideoDescription(vid.description || "");
    setVideoUnlisted(vid.unlisted);
  };

  useEffect(() => {
    fetchVideo();
  }, [params.id]);

  return (
    <main className={`mt-2`}>
      <div className="w-fit mx-auto">
        <ReactShakaPlayer onLoad={(player) => setMainPlayer(player)} uiConfig={{
          'overflowMenuButtons': ['picture_in_picture', 'playback_rate', 'quality'],
          'controlPanelElements': ['play_pause','time_and_duration', 'mute', 'volume', 'spacer', 'overflow_menu', 'fullscreen',]
        }} autoPlay />
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
                <PencilSquareIcon className="w-6 cursor-pointer text-gray-400" onClick={() => setEditModalOpen(true)} />
                <p className="text-sm">•</p>
                <TrashIcon
                  className="w-6 text-red-500 cursor-pointer"
                  onClick={async () => {
                    setDeleteModalOpen(true);
                  }}
                />
              </>
            )}
            <Modal
              open={deleteModalOpen}
              onClose={() => setDeleteModalOpen(false)}
              title="Delete Clip"
              content={
                <>
                  <p className="text-gray-300">Are you sure you want to delete this clip?</p>
                  <p className="text-gray-400 text-md">This action cannot be undone.</p>
                  <div className="flex flex-row space-x-2 w-full justify-end mt-4">
                    <button className="btn btn-outline btn-sm" onClick={() => setDeleteModalOpen(false)}>
                      Cancel
                    </button>
                    <button
                      className="btn btn-error btn-sm"
                      onClick={async () => {
                        await deleteCip(videoDetails.id);
                        router.push("/");
                      }}
                    >
                      Delete
                    </button>
                  </div>
                </>
              }
            />
            <Modal
              open={editModalOpen}
              onClose={() => setEditModalOpen(false)}
              title="Edit Clip"
              content={
                <>
                  <div className="flex flex-col space-y-2">
                    <div className="flex flex-col space-y-2 my-2">
                      <input
                        type="text"
                        className="input"
                        placeholder="Title"
                        value={videoTitle}
                        onChange={(e) => setVideoTitle(e.target.value)}
                      />
                      <textarea
                        className="input"
                        placeholder="Description"
                        value={videoDescription}
                        onChange={(e) => setVideoDescription(e.target.value)}
                      />
                      <label className="label space-x-2 cursor-pointer self-start">
                        <span className="label-text">Unlisted</span>
                        <input
                          type="checkbox"
                          checked={videoUnlisted}
                          onChange={(e) => {
                            setVideoUnlisted(e.target.checked);
                          }}
                          className="checkbox"
                        />
                      </label>
                    </div>
                    <div className="flex flex-row space-x-2 w-full justify-end mt-4">
                      <button className="btn btn-outline btn-sm" onClick={() => setEditModalOpen(false)}>
                        Cancel
                      </button>
                      <button
                        className="btn btn-primary btn-sm"
                        onClick={async () => {
                          await updateClipDetails(videoDetails.id, {
                            title: videoTitle,
                            description: videoDescription,
                            unlisted: videoUnlisted,
                          });
                          setEditModalOpen(false);
                          fetchVideo();
                        }}
                      >
                        Save
                      </button>
                    </div>
                  </div>
                </>
              }
            />
          </div>
        </div>
      )}
    </main>
  );
}
