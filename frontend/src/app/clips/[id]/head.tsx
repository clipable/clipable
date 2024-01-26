import { Clip } from "@/shared/api";
import { headers } from "next/headers";

function staticParams() {}

export const generateStaticParams = process.env.NODE_ENV === "production" ? staticParams : undefined;
export const dynamic = process.env.NODE_ENV === "production" ? "auto" : "force-dynamic";

export default async function Head({ params }: { params: { id?: string } }) {
  const baseItems = [
    <>
      <title>Clipable</title>
      <meta content="width=device-width, initial-scale=1" name="viewport" />
      <meta name="description" content="Easy clip sharing with friends" />
      <link rel="icon" href="/favicon.ico" />
    </>,
  ];

  if (!params.id) {
    return baseItems;
  }

  let data = null;

  try {
    data = await fetch(`http://localhost:3000/api/clips/${params.id}`);
  } catch (e) {
    console.error(e);
  }

  if (!data || !data.ok) {
    return baseItems;
  }

  let video: Clip | null = null;

  try {
    video = await data.json();
  } catch (e) {
    console.error(e);
  }

  if (!video) {
    return baseItems;
  }

  const h = headers();

  const xForwardedHost = h.get("x-forwarded-host") || "localhost:3000";
  const xForwardedProto = h.get("x-forwarded-proto") || "http";

  const proto = xForwardedProto === "https" ? "https" : "http";

  const canonicalVideoPath = `${proto}://${xForwardedHost}/api/clips/${params.id}/dash-stream0.mp4`;
  const canonicalThumbnailPath = `${proto}://${xForwardedHost}/api/clips/${params.id}/thumbnail.jpg`;
  return (
    <>
      {baseItems}
      <meta property="og:title" content={`${video.title} - Clipable`} />
      <meta property="og:image" content={canonicalThumbnailPath} />
      <meta property="og:description" content={video.description} />
      <meta property="og:video" content={canonicalVideoPath} />
      <meta property="og:video:secure_url" content={canonicalVideoPath} />
      <meta property="og:video:type" content="video/mp4" />
      <meta property="og:video:width" content="1280" />
      <meta property="og:video:height" content="720" />
      <meta property="twitter:card" content="player" />
      <meta property="twitter:player" content={canonicalVideoPath} />
      <meta property="twitter:player:width" content="1280" />
      <meta property="twitter:player:height" content="720" />
      <meta property="twitter:image" content={canonicalThumbnailPath} />
      <meta property="twitter:description" content={video.description} />
      <meta property="twitter:title" content={`${video.title} - Clipable`} />
      <meta property="twitter:video:stream" content={canonicalVideoPath} />
      <meta property="twitter:video:stream:content_type" content="video/mp4" />
    </>
  );
}
