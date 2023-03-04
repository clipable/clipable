const API_URL = "/api";

export interface User {
  id: string;
  username: string;
  joined_at: string;
}

export interface Clip {
  created_at: string;
  creator: User;
  description?: string;
  id: string;
  processing: boolean;
  title: string;
  unlisted: boolean;
  views: number;
}

export type ProgressObject = Record<string, number>;

export interface Progress {
  clips: ProgressObject;
}

// Client only
export const getClips = async (): Promise<Clip[]> => {
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

export const getClip = async (videoId: string): Promise<Clip | null> => {
  const response = await fetch(`${API_URL}/clips/${videoId}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (!response.ok) {
    return null;
  }
  return response.json();
};

export const deleteCip = async (videoId: string): Promise<boolean> => {
  const response = await fetch(`${API_URL}/clips/${videoId}`, {
    credentials: "include",
    method: "DELETE",
    headers: {
      "Content-Type": "application/json",
    },
  });

  return response.ok;
};


export const getUsersClips = async (userId: string): Promise<Clip[]> => {
  const response = await fetch(`${API_URL}/users/${userId}/clips`, {
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

export const searchClips = async (query: string): Promise<Clip[]> => {
  const response = await fetch(`${API_URL}/clips/search?query=${query}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
    },
  });
  if (response.status === 204) {
    return [];
  }
  return response.json();
}
