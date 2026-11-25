import { useRef, useState } from "react";

import { useToast } from "../../../../hooks/useToast";
import styles from "../StoreSettings.module.css";

// Logos are stored in localStorage as data URLs, so keep them small.
const MAX_BYTES = 1024 * 1024;

export default function LogoDropZone({ onLogo, tr }) {
  const inputRef = useRef(null);
  const [dragging, setDragging] = useState(false);
  const toast = useToast();

  const readFile = (file) => {
    if (!file) return;
    if (!file.type.startsWith("image/")) return toast.error("Please choose an image file", "يرجى اختيار ملف صورة");
    if (file.size > MAX_BYTES) return toast.error("Logo must be under 1 MB", "يجب أن يكون الشعار أقل من 1 ميغابايت");

    const reader = new FileReader();
    reader.onload = () => onLogo(reader.result);
    reader.readAsDataURL(file);
  };

  return (
    <div
      role="button"
      tabIndex={0}
      className={`${styles.dropZone} ${dragging ? styles.dragging : ""}`}
      onClick={() => inputRef.current.click()}
      onKeyDown={(e) => (e.key === "Enter" || e.key === " ") && inputRef.current.click()}
      onDragOver={(e) => {
        e.preventDefault();
        setDragging(true);
      }}
      onDragLeave={() => setDragging(false)}
      onDrop={(e) => {
        e.preventDefault();
        setDragging(false);
        readFile(e.dataTransfer.files[0]);
      }}
    >
      <i className="fa-solid fa-cloud-arrow-up" />
      <strong>{tr("Drag & drop your logo here", "اسحب وأفلت الشعار هنا")}</strong>
      <span>{tr("or click to browse", "أو انقر للاختيار")}</span>
      <small>{tr("PNG, JPG or SVG · max 1 MB", "PNG أو JPG أو SVG · حتى 1 ميغابايت")}</small>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        hidden
        onChange={(e) => {
          readFile(e.target.files[0]);
          e.target.value = "";
        }}
      />
    </div>
  );
}
