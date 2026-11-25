import { useContext, useState } from "react";

import { AuthContext } from "../../../context/AuthContext";
import { useLang } from "../../../hooks/useLang";
import { useToast } from "../../../hooks/useToast";
import { initials } from "../../../utils/format";
import styles from "./AddCommentForm.module.css";

// Review form for the signed-in user. onSubmit(comment, rating) posts to the API.
function AddCommentForm({ onSubmit, onClose }) {
  const { user } = useContext(AuthContext);
  const { tr, dir } = useLang();
  const toast = useToast();
  const [rating, setRating] = useState(5);
  const [comment, setComment] = useState("");
  const [error, setError] = useState("");
  const [saving, setSaving] = useState(false);

  const submit = async (e) => {
    e.preventDefault();
    if (!comment.trim()) return setError(tr("Please write your review", "يرجى كتابة رأيك"));
    setSaving(true);
    try {
      await onSubmit(comment.trim(), rating);
      toast.success("Thanks for your review!", "شكرًا على تقييمك!");
      onClose();
    } catch (err) {
      setError(err.status === 409 ? tr("You already reviewed this product", "لقد قيّمت هذا المنتج من قبل") : err.message);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className={styles.overlay} onMouseDown={onClose} dir={dir}>
      <form className={styles.form} onSubmit={submit} onMouseDown={(e) => e.stopPropagation()}>
        <div className={styles.header}>
          {user?.image ? <img src={user.image} alt="" /> : <span className={styles.avatar}>{initials(user?.userFullName)}</span>}
          <strong>{user?.userFullName}</strong>
          <button type="button" className={styles.close} onClick={onClose} aria-label={tr("Close", "إغلاق")}>
            <i className="fa-solid fa-xmark"></i>
          </button>
        </div>

        <div className={styles.stars} role="radiogroup" aria-label={tr("Rating", "التقييم")}>
          {[1, 2, 3, 4, 5].map((n) => (
            <button
              type="button"
              key={n}
              role="radio"
              aria-checked={rating === n}
              aria-label={`${n}`}
              className={n <= rating ? styles.on : ""}
              onClick={() => setRating(n)}
            >
              <i className="fa-solid fa-star"></i>
            </button>
          ))}
        </div>

        <textarea
          value={comment}
          maxLength={2000}
          placeholder={tr("What did you like or dislike?", "ما الذي أعجبك أو لم يعجبك؟")}
          onChange={(e) => {
            setComment(e.target.value);
            setError("");
          }}
        />
        {error && <p className={styles.error}>{error}</p>}

        <div className={styles.actions}>
          <button type="button" className={styles.cancel} onClick={onClose}>
            {tr("Cancel", "إلغاء")}
          </button>
          <button type="submit" className={styles.submit} disabled={saving}>
            {saving ? <i className="fa-solid fa-spinner fa-spin"></i> : tr("Submit", "إرسال")}
          </button>
        </div>
      </form>
    </div>
  );
}

export default AddCommentForm;
