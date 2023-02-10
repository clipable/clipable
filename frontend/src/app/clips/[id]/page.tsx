"use client";

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import { useState, useEffect, useRef } from "react";
import Link from "next/link";
import ShakaPlayer from "shaka-player-react";
import "shaka-player-react/dist/controls.css";
// import { ControlBar } from "./controlbar";
// import "../../../styles/controlbar.css";

import dashjs from "dashjs";

const inter = Inter({ subsets: ["latin"] });

interface SkakaPlayerProperties {
  player: any,
  ui: any,
  videoElement: HTMLVideoElement,
}

export default function Page({ params }: { params: { id: string } }) {
  const loggedIn = false;



  const videoRef = useRef(null);
  // useEffect(() => {
  //   const {
  //     player,
  //     ui,
  //     videoElement
  //   } = controllerRef.current as unknown as SkakaPlayerProperties;

  //   async function loadAsset() {
  //     // Load an asset.
  //     await player?.load(`http://localhost:8080/api/clips/${params.id}/dash.mpd`);

  //     // Trigger play.
  //     // videoElement.play();
  //   }
  //   console.log(player)
  //   loadAsset();
  // }, [controllerRef, params.id]);
  console.log(videoRef);
  console.log(params.id);

  return (
    <main className="bg-white dark:bg-black max-w-[70%] mt-6 mx-auto">
      <ShakaPlayer src={`http://localhost:8080/api/clips/${params.id}/dash.mpd`} ref={videoRef} autoPlay={true} chromeless={false}/>
    </main>
  );
}
