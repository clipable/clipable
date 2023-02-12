"use client";

import clsx from "clsx";
import { Video } from "@/shared/api";

interface Props {
  video: Video;
}

export default function VideoCard({ video }: Props) {
  const figureClassname = clsx({
    "flex-auto grow self-center pt-8 glass w-full h-full": video.processing,
  });

  const cardBodyClassname = clsx({
    "card-body justify-end": video.processing,
    "card-body": !video.processing,
  });

  return (
    <div className="card card-compact w-96 h-full bg-base-100 shadow-xl min-w-[28rem] min-h-[20rem]">
      <figure className={figureClassname}>
        {video.processing ? <div>Processing!</div> : <img src={`/api/clips/${video.id}/thumbnail.jpg`} />}
      </figure>
      <div className={cardBodyClassname}>
        <div className="flex-row flex text-xl">
          <div className="container flex card-title">{video.title}</div>
          <div className="container flex card-title justify-end">{video.views}</div>
        </div>
        <p className="grow-0">{video.description}</p>
      </div>
    </div>
  );
}
