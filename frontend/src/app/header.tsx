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
            <span className="btn btn-ghost normal-case text-xl">Clipable</span>
          </Link>
          <div className="space-x-4">
            {userContext.loggedIn && !userContext.loading && (
              <Link href="/upload">
                <button className="btn btn-primary btn-sm">Upload</button>
              </Link>
            )}
            {!userContext.loggedIn && !userContext.loading && (
              <Link href="/login">
                <button className="btn btn-primary btn-sm">Login</button>
              </Link>
            )}
            {userContext.loggedIn && !userContext.loading && (
              <button
                className="btn btn-outline btn-sm"
                onClick={async () => {
                  await logout();
                  window.location.reload();
                }}
              >
                Logout
              </button>
            )}
          </div>
        </div>
      </nav>
    </header>
  );
}
