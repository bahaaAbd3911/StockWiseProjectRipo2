import { useLang } from "../../../hooks/useLang";
import Button from "./Button";
import Modal from "./Modal";

export default function ConfirmDialog({ title, message, confirmLabel, onConfirm, onCancel }) {
  const { tr } = useLang();

  return (
    <Modal
      title={title}
      onClose={onCancel}
      footer={
        <>
          <Button variant="ghost" onClick={onCancel}>
            {tr("Cancel", "إلغاء")}
          </Button>
          <Button variant="danger" icon="fa-solid fa-trash" onClick={onConfirm}>
            {confirmLabel ?? tr("Delete", "حذف")}
          </Button>
        </>
      }
    >
      <p style={{ margin: 0, lineHeight: 1.6 }}>{message}</p>
    </Modal>
  );
}
