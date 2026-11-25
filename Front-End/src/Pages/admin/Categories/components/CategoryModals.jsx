import { ConfirmDialog } from "../../../../components/admin/ui";
import CategoryFormModal from "./CategoryFormModal";

// Form + delete confirmation driven by useCategoryCrud().
export default function CategoryModals({ crud, isAr, tr }) {
  const { editing, deleting } = crud;
  const count = deleting ? crud.countFor(deleting) : 0;
  const name = deleting && (isAr ? deleting.nameAr : deleting.name);

  return (
    <>
      {editing && (
        <CategoryFormModal
          category={editing.id !== undefined ? editing : null}
          categories={crud.categories}
          tr={tr}
          onSave={crud.save}
          onClose={crud.closeForm}
        />
      )}

      {deleting && (
        <ConfirmDialog
          title={tr("Delete category", "حذف التصنيف")}
          message={
            count
              ? tr(
                  `Delete "${name}"? Its ${count} products will be deleted too.`,
                  `حذف "${name}"؟ سيتم حذف ${count} منتج تابع له أيضًا.`,
                )
              : tr(`Delete "${name}"?`, `حذف "${name}"؟`)
          }
          onConfirm={crud.confirmDelete}
          onCancel={crud.cancelDelete}
        />
      )}
    </>
  );
}
