"use client";

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";
// import { ControlBar } from "./controlbar";
// import "../../../styles/controlbar.css";

import dashjs from "dashjs";

const inter = Inter({ subsets: ["latin"] });

export default function Page({ params }: { params: { id: string } }) {
  const loggedIn = false;

  const videoRef = useRef<HTMLVideoElement>(null);

  const [player, setPlayer] = useState<dashjs.MediaPlayerClass | null>(null);

  const [volume, setVolume] = useState<number>(50);

  //   const [videos, setVideos] = useState<Videos[]>([]);

  //   useEffect(() => {
  //     const getVids = async () => {
  //       const vids = await getVideos();
  //       setVideos(vids);
  //     };
  //     getVids();
  //   }, []);

  useEffect(() => {
    if (videoRef.current) {
      const player = dashjs.MediaPlayer().create();
      player.initialize(videoRef.current, `http://localhost:8080/api/clips/${params.id}/dash.mpd`, true);

      player.attachView(videoRef.current);

      setPlayer(player);
      // const controlbar = new ControlBar(player) as any;
      //Player is instance of Dash.js MediaPlayer;
      // controlbar.initialize();
    }
  }, [params.id]);

  const changeVolume = (change: React.ChangeEvent<HTMLInputElement>) => {
    const value = parseInt(change.target.value, 10);
    if (isNaN(value)) return;

    setVolume(value);

    console.log(player?.getVolume());

    // volume is in range 0-1
    // so we need to divide by 100

    const volume = value / 100;

    player?.setVolume(volume);
  }

  const playPauseVideo = () => {
    player?.isPaused() ? player?.play() : player?.pause();
  };

  console.log(params.id);

  return (
    <main className="h-screen bg-white dark:bg-black">
      <header className="bg-gray-900">
        <nav className="flex px-2 lg:px-8" aria-label="Top">
          <div className="flex w-full grow justify-between border-b border-indigo-500 py-1 lg:border-none">
            <div className="flex items-center">
              <Link href="/">
                <span className="dark:text-white font-bold">Clipable</span>
              </Link>
            </div>
            <div className="ml-10 space-x-4">
              {loggedIn ? (
                // @TODO: Add a my clips button?
                // @TODO: Abstact these to a buttom component
                // @TODO: Maybe use DaisyUI?
                <a
                  href="#"
                  className="inline-block rounded-md border border-transparent bg-white py-1 px-2 text-base font-medium text-indigo-600 hover:bg-indigo-50"
                >
                  Sign out
                </a>
              ) : (
                <a
                  href="#"
                  className="inline-block rounded-md border border-transparent bg-indigo-500 py-1 px-2 text-base font-medium text-white hover:bg-opacity-75"
                >
                  Sign in
                </a>
              )}
            </div>
          </div>
        </nav>
      </header>
      <div>
        <video slot="media" controls ref={videoRef} preload="auto" autoPlay className="w-4/5 mx-auto pt-10" />
        <button className="btn" onClick={playPauseVideo} />
        <input type="range" min="0" max="100" value={volume} onChange={changeVolume} className="range range-xs" />
      </div>
    </main>
  );
}
