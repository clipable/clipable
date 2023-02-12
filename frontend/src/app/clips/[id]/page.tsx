"use client";

import ShakaPlayer from "shaka-player-react";
import "shaka-player-react/dist/controls.css";

export default function Page({ params }: { params: { id: string } }) {
  return (
    <main className="bg-white dark:bg-black max-w-[70%] mt-6 mx-auto">
      <ShakaPlayer src={`http://localhost:8080/api/clips/${params.id}/dash.mpd`} autoPlay={true} />
    </main>
  );
}
