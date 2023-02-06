"use client";

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";
import { ControlBar } from "./controlbar";
import "../../../styles/controlbar.css";

import dashjs from "dashjs";

const inter = Inter({ subsets: ["latin"] });

export default function Page({ params }: { params: { id: string } }) {
  const loggedIn = false;

  const videoRef = useRef<HTMLVideoElement>(null);

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

      player.attachView(videoRef.current)
      const controlbar = new ControlBar(player) as any;
      //Player is instance of Dash.js MediaPlayer;
      controlbar.initialize();
    }
  }, [params.id]);

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
        <video slot="media" controls={false} ref={videoRef} preload="auto" autoPlay={true} className="w-4/5 mx-auto pt-10" />
        <div id="videoController" className="video-controller unselectable">
          <div id="playPauseBtn" className="btn-play-pause" title="Play/Pause">
            <span id="iconPlayPause" className="icon-play"></span>
          </div>
          <span id="videoTime" className="time-display">00:00:00</span>
          <div id="fullscreenBtn" className="btn-fullscreen control-icon-layout" title="Fullscreen">
            <span className="icon-fullscreen-enter"></span>
          </div>
          <div id="bitrateListBtn" className="control-icon-layout" title="Bitrate List">
            <span className="icon-bitrate"></span>
          </div>
          <input type="range" id="volumebar" className="volumebar" value="1" min="0" max="1" step=".01" />
          <div id="muteBtn" className="btn-mute control-icon-layout" title="Mute">
            <span id="iconMute" className="icon-mute-off"></span>
          </div>
          <div id="trackSwitchBtn" className="control-icon-layout" title="A/V Tracks">
            <span className="icon-tracks"></span>
          </div>
          <div id="captionBtn" className="btn-caption control-icon-layout" title="Closed Caption">
            <span className="icon-caption"></span>
          </div>
          <span id="videoDuration" className="duration-display">00:00:00</span>
          <div className="seekContainer">
            <div id="seekbar" className="seekbar seekbar-complete">
              <div id="seekbar-buffer" className="seekbar seekbar-buffer"></div>
              <div id="seekbar-play" className="seekbar seekbar-play"></div>
            </div>
          </div>
          <div id="thumbnail-container" className="thumbnail-container">
            <div id="thumbnail-elem" className="thumbnail-elem"></div>
            <div id="thumbnail-time-label" className="thumbnail-time-label"></div>
          </div>
        </div>
      </div>
    </main>
  );
}
