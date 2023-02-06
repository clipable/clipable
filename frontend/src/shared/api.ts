import { cookies } from 'next/headers'

const API_URL = 'http://localhost:8080/api';

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

export const getVideos = async (): Promise<Videos[]> => {
  // const nextCookies = cookies()
  // const token = nextCookies.get('webserver')
  // if (!token) {
  //   return Promise.reject()
  // }
  const response = await fetch(`${API_URL}/clips`, {
    credentials: "include",
    headers: {
      // cookie: token,
      'Content-Type': 'application/json'
    }
  });
  console.log(`RESPONSE CODE: ${response.status}`)
  console.log(await response.text())
  return response.json();
}