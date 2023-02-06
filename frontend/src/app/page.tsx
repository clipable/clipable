'use client';

import { getVideos, Videos } from "@/shared/api";
import { Inter } from "@next/font/google";
import { useState, useEffect } from "react";

const inter = Inter({ subsets: ["latin"] });

export default function Home() {

  const loggedIn = false;

  const [videos, setVideos] = useState<Videos[]>([])

  useEffect(() => {
    const getVids = async () => {
      const vids = await getVideos();
      setVideos(vids);
    }
    getVids()
  }, [])

  console.log(videos);

  return (
    <main className="h-screen bg-white dark:bg-black">
      <header className="bg-gray-900">
        <nav className="flex px-2 lg:px-8" aria-label="Top">
          <div className="flex w-full grow justify-between border-b border-indigo-500 py-1 lg:border-none">
            <div className="flex items-center">
              <a href="#">
                <span className="dark:text-white font-bold">Clipable</span>
              </a>
            </div>
            <div className="ml-10 space-x-4">
              {loggedIn ? (
                // @TODO: Add a my clips button?
                // @TODO: Abstact these to a buttom component
                // @TODO: Maybe use DaisyUI?
              <a
                href="#"
                className="inline-block rounded-md border border-transparent bg-white py-1 px-2 text-base font-medium text-indigo-600 hover:bg-indigo-50"
              >
                Sign out
              </a>
              ) : (
              <a
                href="#"
                className="inline-block rounded-md border border-transparent bg-indigo-500 py-1 px-2 text-base font-medium text-white hover:bg-opacity-75"
              >
                Sign in
              </a>

              )}
            </div>
          </div>
        </nav>
      </header>
      <div className="grid">
        
      </div>
    </main>
  );
}
