import { useLang } from "../../../hooks/useLang";
import { Button, Modal } from "../ui";
import { useProductForm } from "./useProductForm";
import ProductNameFields from "./ProductNameFields";
import ProductPricingFields from "./ProductPricingFields";
import ProductMediaFields from "./ProductMediaFields";
import ProductDescriptionFields from "./ProductDescriptionFields";
import ColorsEditor from "./ColorsEditor";

// Add/edit product. Pass `lockedCategory` to pre-select and lock the category.
export default function ProductFormModal({ product, categories, lockedCategory, onSave, onClose }) {
  const { tr, isAr } = useLang();
  const { form, errors, update, submit } = useProductForm({
    product,
    defaultCategory: lockedCategory,
    categories,
    tr,
    onSave,
  });
  const shared = { form, errors, update, tr, isAr };

  return (
    <Modal
      wide
      title={product ? tr("Edit Product", "تعديل المنتج") : tr("Add Product", "إضافة منتج")}
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>{tr("Cancel", "إلغاء")}</Button>
          <Button type="submit" form="productForm" icon="fa-solid fa-check">
            {tr("Save Product", "حفظ المنتج")}
          </Button>
        </>
      }
    >
      <form id="productForm" onSubmit={submit} noValidate>
        <ProductNameFields {...shared} categories={categories} lockCategory={Boolean(lockedCategory)} />
        <ProductPricingFields {...shared} />
        <ProductMediaFields {...shared} />
        <ProductDescriptionFields {...shared} />
        <ColorsEditor
          colors={form.availableColors}
          error={errors.availableColors}
          tr={tr}
          onChange={(colors) => update("availableColors", colors)}
        />
      </form>
    </Modal>
  );
}
