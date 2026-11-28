import { useState } from "react";

import { saveOrderNote } from "../../../api/admin";
import { useApiAction } from "../../../hooks/useApiAction";
import { formatDate } from "../../../utils/format";
import { Button, FormField } from "../ui";
import styles from "./orders.module.css";

const MAX = 2000; // same limit as the backend (orders.admin_note)

// The store's private note on an order ("deliver after 5 pm", "gift wrap").
// Only admins see it; it never reaches the customer.
export default function OrderNoteCard({ order, isAr, tr, onSaved }) {
  const { run, busy } = useApiAction();
  const [note, setNote] = useState(order.adminNote ?? "");
  const dirty = note.trim() !== (order.adminNote ?? "");

  const save = () =>
    run(async () => {
      await saveOrderNote(order.id, note);
      await onSaved?.();
    }, ["Note saved", "تم حفظ الملاحظة"]);

  return (
    <section className={styles.section}>
      <h3>
        <i className="fa-solid fa-lock" /> {tr("Private note", "ملاحظة خاصة")}
      </h3>
      <FormField
        as="textarea"
        name="adminNote"
        rows={3}
        maxLength={MAX}
        value={note}
        placeholder={tr(
          'Only admins see this, e.g. "deliver after 5 pm"',
          'يراها المشرفون فقط، مثل "التوصيل بعد الخامسة"',
        )}
        hint={
          order.adminNoteUpdatedAt
            ? `${tr("Last changed", "آخر تعديل")} ${formatDate(order.adminNoteUpdatedAt, isAr, true)}`
            : `${note.length}/${MAX}`
        }
        onChange={(e) => setNote(e.target.value)}
      />
      <div className={styles.noteActions}>
        <Button
          icon="fa-solid fa-floppy-disk"
          disabled={!dirty || busy}
          onClick={save}
        >
          {tr("Save note", "حفظ الملاحظة")}
        </Button>
      </div>
    </section>
  );
}
