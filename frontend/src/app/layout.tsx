import '@/styles/globals.scss';
import Link from 'next/link';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      {/*
        <head /> will contain the components returned by the nearest parent
        head.tsx. Find out more at https://beta.nextjs.org/docs/api-reference/file-conventions/head
      */}
      <head />
      <body>
        <header className="navbar bg-base-300">
          <nav className="flex px-2 lg:px-8" aria-label="Top">
            <div className="flex w-full grow justify-between border-b border-indigo-500 py-1 lg:border-none">
              <div className="flex items-center">
                <Link href="/">
                  <span className="dark:text-white font-bold">Clipable</span>
                </Link>
              </div>
            </div>
          </nav>
        </header>
        {children}
      </body>
    </html>
  );
}
