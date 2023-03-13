"use client";

import clsx from "clsx";
import { Clip } from "@/shared/api";
import { formatViewsCount } from "./views-formatter";
import { formatDate } from "./date-formatter";
import ClipCardImage from "./clip-card-image";
import Link from "next/link";

interface Props {
  video: Clip;
  progress?: number;
}

export default function ClipCard({ video, progress }: Props) {
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
        <ClipCardImage video={video} progress={progress} />
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
          <div className="flex flex-row w-fit space-x-2 text-gray-400 text-base font-thin items-baseline">
            <p>
              {formatViewsCount(video.views)} view{video.views === 1 ? "" : "s"}
            </p>
            <p className="text-sm">•</p>
            <p>{formatDate(video.created_at)}</p>
          </div>
          {video.unlisted && <div className="badge badge-outline mt-1">unlisted</div>}
        </div>
      </div>
    </div>
  );
}
