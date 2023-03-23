"use client";

import Link from "next/link";

export default function Footer() {
  return (
    <footer className="flex fixed bottom-0 items-center place-content-center pb-1 pt-1 w-full bg-base-300 text-gray-400">
      &copy; 2023 Clipable. All rights reserved.{" "}
      <a href="https://github.com/clipable/clipable" target="_blank">
        <div className="text-blue-400"> &nbsp; Open Source on GitHub</div>
      </a>
    </footer>
  );
}
4;
