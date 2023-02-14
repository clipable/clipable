"use client";

interface Props {
  videoId: string;
  progress?: number;
}

/**
 * Renders the thumbnail of a video or the progress of the video if it is being processed
 */
export default function VideoCardImage({ progress, videoId }: Props) {
  
  if (!progress) {
    return <img src={`/api/clips/${videoId}/thumbnail.jpg`} />
  }

  if (progress === -1) {
    return <div style={{ userSelect: "none" }}>Queued</div>
  }

  // @TODO: Since we have a thumbnail here we can maybe add some opacity to it and show the progress on top of it
  return (
    <div className="radial-progress" style={{"--value":progress} as any}>{`${progress}%`}</div>
  )
}
