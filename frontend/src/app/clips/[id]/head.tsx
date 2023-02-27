import { Video, getVideo } from "@/shared/api";

async function getVideoDetails(id: string) {
  const resp = await fetch(
    `http://localhost:8080/api/clips/${id}`, {
        credentials: "include",
        headers: {
          "Content-Type": "application/json",
        },
    }
  );
    const videoDetails = await resp.json();
  return videoDetails;
}

export default async function Head({ params }: { params: { id: string } }) {
    const videoDetails = await getVideoDetails(params.id);
    return (
      <>
        <title>{`${videoDetails?.title} | Clipable`}</title>
      </>
    );
  }
  