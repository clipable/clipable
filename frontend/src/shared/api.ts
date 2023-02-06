import { cookies } from "next/headers";

const API_URL = "http://localhost:8080/api";

interface User {
  id: string;
  username?: string;
  email?: string;
  joined_at: string;
}

export interface Videos {
  id: string;
  title: string;
  description?: string;
  created_at: string;
  creator: User;
}

// Client only
export const getVideos = async (): Promise<Videos[]> => {
  const response = await fetch(`${API_URL}/clips`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  return response.json();
};
