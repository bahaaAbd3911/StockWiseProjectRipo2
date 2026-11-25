import { ConfirmDialog } from "../../../../components/admin/ui";
import UserFormModal from "./UserFormModal";

// Form + delete confirmation driven by useUserCrud(). onDeleted runs after a delete.
export default function UserModals({ crud, isAr, tr, onDeleted }) {
  const { editing, deleting } = crud;

  return (
    <>
      {editing && (
        <UserFormModal
          user={editing.userID !== undefined ? editing : null}
          users={crud.users}
          isAr={isAr}
          tr={tr}
          onSave={crud.save}
          onClose={crud.closeForm}
        />
      )}

      {deleting && (
        <ConfirmDialog
          title={tr("Delete user", "حذف المستخدم")}
          message={tr(
            `Are you sure you want to delete ${deleting.userFullName}?`,
            `هل أنت متأكد من حذف ${deleting.userFullName}؟`,
          )}
          onConfirm={() => {
            crud.confirmDelete();
            onDeleted?.();
          }}
          onCancel={crud.cancelDelete}
        />
      )}
    </>
  );
}
