"use client";
import { getUser, User } from "@/shared/api";
import React, { createContext, useEffect, useState } from "react";

interface UserContext {
  user: User | undefined;
  loggedIn: boolean;
  loading: boolean;
  reload: () => void;
}

export const UserContext = createContext<UserContext>({
  user: undefined,
  loggedIn: false,
  loading: true,
  reload: () => {},
});

export default function UserProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | undefined>();
  const [loading, setLoading] = useState<boolean>(true);
  const [shouldReload, setShouldReload] = useState<boolean>(false);

  useEffect(() => {
    const fetchUser = async () => {
      setLoading(true);
      const user = await getUser();
      if (user) {
        setUser(user);
      }
      setLoading(false);
    };
    fetchUser();
  }, [shouldReload]);

  return (
    <UserContext.Provider
      value={{
        loading,
        user: user,
        loggedIn: !!user,
        reload: () => {
          setShouldReload(!shouldReload);
        },
      }}
    >
      {children}
    </UserContext.Provider>
  );
}
