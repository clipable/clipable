"use client";

import dynamic from 'next/dynamic';

const ShakaPlayer = dynamic(
  () => import('shaka-player-react'), 
  { ssr: false },
);

import "shaka-player-react/dist/controls.css";

export default function Page({ params }: { params: { id: string } }) {
  return (
    <main className="bg-white dark:bg-black max-w-[70%] mt-6 mx-auto">
      <ShakaPlayer src={`/api/clips/${params.id}/dash.mpd`} autoPlay={true} />
    </main>
  );
}
