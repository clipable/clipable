import clsx from "clsx";
import { useRef } from "react";
import { useClickOutside } from "./use-click-outside";

interface Props {
  open: boolean;
  onClose: () => void;
  title: string;
  content: React.ReactNode;
}

function Modal({ open, onClose, title, content }: Props) {
  const modalRef = useRef<HTMLDivElement>(null);
  useClickOutside({
    ref: modalRef,
    callback: () => onClose(),
  });

  const modalClassName = clsx("modal cursor-pointer", open ? "modal-open" : "");

  return (
    <div className={modalClassName}>
      <label className="modal-box relative" htmlFor="" ref={modalRef as any}>
        <h3 className="text-lg font-bold">{title}</h3>
        {content}
      </label>
    </div>
  );
}

export default Modal;
