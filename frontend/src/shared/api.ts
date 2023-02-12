import { cookies } from "next/headers";

const API_URL = "/api";

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
  views: number;
  processing: boolean;
}

// Client only
export const getVideos = async (): Promise<Videos[]> => {
  const response = await fetch(`${API_URL}/clips`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (response.status === 204) {
    return [];
  }
  return response.json();
};

export const getUser = async (): Promise<User | undefined> => {
  const response = await fetch(`${API_URL}/users/me`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (response.ok) {
    return response.json();
  }
};

export const register = async (username: string, password: string): Promise<Response> => {
  const response = await fetch(`${API_URL}/auth/register`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });
  return response;
};

export const login = async (username: string, password: string): Promise<boolean> => {
  const response = await fetch(`${API_URL}/auth/login`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ username, password }),
  });
  return response.ok;
};

export const logout = async (): Promise<boolean> => {
  const response = await fetch(`${API_URL}/auth/logout`, {
    method: "POST",
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  return response.ok;
};
