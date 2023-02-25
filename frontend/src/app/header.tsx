"use client";

import { UserContext } from "@/context/user-context";
import { logout } from "@/shared/api";
import Link from "next/link";
import { useContext } from "react";

export default function Header() {
  const userContext = useContext(UserContext);
  return (
    <header className="navbar bg-base-300">
      <nav className="flex w-full px-2 lg:px-8" aria-label="Top">
        <div className="flex items-center w-full grow justify-between border-b border-indigo-500 py-1 lg:border-none">
          <Link href="/">
            <span className="btn btn-ghost normal-case text-xl library">Clipable</span>
          </Link>
          <div className="space-x-4">
            {userContext.loggedIn && !userContext.loading && (
              <Link href="/upload">
                <button className="btn btn-outline btn-sm">Upload</button>
              </Link>
            )}
            {!userContext.loggedIn && !userContext.loading ? (
              <Link href="/login">
                <button className="btn btn-primary btn-sm">Login</button>
              </Link>
            ) : (
              <ul className="menu menu-compact menu-horizontal">
                <li tabIndex={0}>
                  <a className="uppercase font-semibold">
                    {userContext.user?.username}
                    <svg
                      className="fill-current"
                      xmlns="http://www.w3.org/2000/svg"
                      width="20"
                      height="20"
                      viewBox="0 0 24 24"
                    >
                      <path d="M7.41,8.58L12,13.17L16.59,8.58L18,10L12,16L6,10L7.41,8.58Z" />
                    </svg>
                  </a>
                  <ul className="bg-base-100 w-full">
                    {userContext.loggedIn && !userContext.loading && (
                      <>
                        <li className="hover-bordered">
                          <Link href={`users/${userContext.user?.id}`}>
                            <p>My Clips</p>
                          </Link>
                        </li>
                        <li
                          className="hover:text-red-500 hover:border-red-500 hover:border-l-4 border-l-4 border-transparent"
                          onClick={async () => {
                            await logout();
                            window.location.reload();
                          }}
                        >
                          <a>Logout</a>
                        </li>
                      </>
                    )}
                  </ul>
                </li>
              </ul>
            )}
          </div>
        </div>
      </nav>
    </header>
  );
}
