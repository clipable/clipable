"use client";

import { useEffect, useState } from "react";
import{useSpring, animated}from"react-spring";
import { Video } from "@/shared/api";

interface Props {
  video: Video;
  progress?: number;
}

/**
 * Renders the thumbnail of a video or the progress of the video if it is being processed
 */
export default function VideoCardImage({ progress, video }: Props) {

  const [oldVal, setOldVal] = useState<number>(0);

  const barvalue = useSpring({
    from: { "--value": oldVal },
    config: { duration: 1000 },
    to: { "--value": progress ?? oldVal },
  });

  useEffect(() => {
    if (progress) {
      setOldVal(progress);
    }
  }, [progress]);

  if (!video.processing) {
    return <img src={`/api/clips/${video.id}/thumbnail.jpg`} />
  }

  if (progress === -1) {
    return <div style={{ userSelect: "none" }}>Queued</div>
  }

  if (progress === undefined) {
    return <div style={{ userSelect: "none" }}>Loading...</div>
  }

  // @TODO: Since we have a thumbnail here we can maybe add some opacity to it and show the progress on top of it
  return (
    <animated.div  className="radial-progress select-none" style={barvalue as any}>
      {barvalue["--value"].to((p) => {
        return p.toFixed(0) + "%"
      })}
      </animated.div>
  )
}
