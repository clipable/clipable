"use client";

import clsx from "clsx";
import { Video } from "@/shared/api";
import { formatViewsCount } from "./views-formatter";
import { formatDate } from "./date-formatter";
import VideoCardImage from "./video-card-image";
import Link from "next/link";

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
    <div className="card card-compact w-96 h-full bg-base-100 min-w-[28rem] min-h-[20rem]">
      <figure className={figureClassname}>
        <VideoCardImage video={video} progress={progress} />
      </figure>
      <div
        className={cardBodyClassname}
        style={{
          padding: 0,
        }}
      >
        <div className="flex flex-col text-xl">
          <div
            className="truncate card-title"
            style={{
              marginBottom: "0.1rem",
            }}
          >
            {video.title}
          </div>
          <Link href={`/users/${video.creator.id}`}>
            <div className="truncate text-stone-400 text-base font-thin hover:text-stone-200">
              {video.creator.username}
            </div>
          </Link>
          <div className="flex flex-row w-fit space-x-2 text-gray-400 text-base font-thin">
            <p>
              {formatViewsCount(video.views)} view{video.views === 1 ? "" : "s"}
            </p>
            <p>•</p>
            <p>{formatDate(video.created_at)}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
