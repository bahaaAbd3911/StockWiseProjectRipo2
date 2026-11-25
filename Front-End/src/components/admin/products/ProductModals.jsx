import { useLang } from "../../../hooks/useLang";
import { ConfirmDialog } from "../ui";
import ProductFormModal from "./ProductFormModal";

// Renders the product form / delete confirmation driven by useProductCrud().
export default function ProductModals({ crud, lockedCategory }) {
  const { tr, isAr } = useLang();
  const { editing, deleting } = crud;

  return (
    <>
      {editing && (
        <ProductFormModal
          product={editing.id !== undefined ? editing : null}
          categories={crud.categories}
          lockedCategory={lockedCategory}
          onSave={crud.save}
          onClose={crud.closeForm}
        />
      )}

      {deleting && (
        <ConfirmDialog
          title={tr("Delete product", "حذف المنتج")}
          message={tr(
            `Are you sure you want to delete "${deleting.name}"? This cannot be undone.`,
            `هل أنت متأكد من حذف "${isAr ? deleting.nameAr : deleting.name}"؟ لا يمكن التراجع عن ذلك.`,
          )}
          onConfirm={crud.confirmDelete}
          onCancel={crud.cancelDelete}
        />
      )}
    </>
  );
}
