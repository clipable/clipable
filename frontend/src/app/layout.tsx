import UserContext from "@/context/user-context";
import "@/styles/globals.scss";
import Header from "./header";

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      {/*
        <head /> will contain the components returned by the nearest parent
        head.tsx. Find out more at https://beta.nextjs.org/docs/api-reference/file-conventions/head
      */}
      <head />
      <body>
        <UserContext>
          <Header />
          {children}
        </UserContext>
      </body>
    </html>
  );
}
