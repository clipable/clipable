"use client";

import clsx from "clsx";
import { Video } from "@/shared/api";
import { formatViewsCount } from "./views-formatter";
import { formatDate } from "./date-formatter";
import VideoCardImage from "./video-card-image";


interface Props {
  video: Video;
  progress?: number;
}

export default function VideoCard({ video, progress }: Props) {
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
        <VideoCardImage video={video} progress={progress} />
      </figure>
      <div className={cardBodyClassname}>
        <div className="flex flex-col text-xl">
          <div className="truncate card-title">{video.title}</div>
          <div className="flex flex-row w-full justify-between space-x-2">
            <div className="mt-2 text-stone-400 text-base">{formatViewsCount(video.views)} view{video.views === 1 ? '' : 's'}</div>
            <div className="mt-2 text-stone-400 text-base">{formatDate(video.created_at)}</div>
          </div>
        </div>
      </div>
    </div>
  );
}
