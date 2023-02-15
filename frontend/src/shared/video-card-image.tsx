"use client";

import { useEffect, useState } from "react";
import{useSpring, animated}from"react-spring";

interface Props {
  videoId: string;
  progress?: number;
}

/**
 * Renders the thumbnail of a video or the progress of the video if it is being processed
 */
export default function VideoCardImage({ progress, videoId }: Props) {

  const [oldVal, setOldVal] = useState<number>(0);

  const barvalue = useSpring({
    from: { progress: oldVal },
    config: { duration: 1500 },
    to: { progress: progress ?? oldVal },
  });

  useEffect(() => {
    if (progress) {
      setOldVal(progress);
    }
  }, [progress]);

  if (!progress) {
    return <img src={`/api/clips/${videoId}/thumbnail.jpg`} />
  }

  if (progress === -1) {
    return <div style={{ userSelect: "none" }}>Queued</div>
  }

  // @TODO: Since we have a thumbnail here we can maybe add some opacity to it and show the progress on top of it
  return (
    <animated.div className="radial-progress" style={{"--value":progress} as any}>
      {barvalue.progress.to((p) => p.toFixed(2) + "%" )}
      </animated.div>
  )
}
