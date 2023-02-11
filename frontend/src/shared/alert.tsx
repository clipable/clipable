"use client";

import clsx from "clsx";

interface Props {
  type: "info" | "success" | "error";
  message: string;
}

export default function Alert({ type, message }: Props) {
  const className = clsx({
    alert: true,
    "alert-info": type === "info",
    "alert-success": type === "success",
    "alert-error": type === "error",
    "shadow-lg": true,
  });

  return (
    <div className={className}>
      <div>
        <span>{message}</span>
      </div>
    </div>
  );
}
