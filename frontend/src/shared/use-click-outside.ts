import React from "react";

interface Props {
  ref: React.RefObject<any>;
  callback: () => void;
}

export const useClickOutside = ({ ref, callback }: Props) => {
  const handleClick = (e: any) => {
    if (ref.current && !ref.current.contains(e.target)) {
      callback();
    }
  };
  React.useEffect(() => {
    document.addEventListener("click", handleClick);
    return () => {
      document.removeEventListener("click", handleClick);
    };
  });
};
