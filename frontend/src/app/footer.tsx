"use client";

import Link from "next/link";

export default function Footer() {
  return (
    <footer className="footer fixed bottom-0 bg-base-300">
      <nav className="flex px-2 lg:px-8" aria-label="Bottom">
        <div>Clipable - A project for self hosting videos</div>
        <div>View the source code on</div>
        <Link href="https://github.com/clipable/clipable">
          <strong>GitHub</strong>
        </Link>
      </nav>
    </footer>
  );
}
