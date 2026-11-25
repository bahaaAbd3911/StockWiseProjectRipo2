import { useState } from "react";

import * as catalogApi from "../../../api/catalog";
import { useCatalog } from "../../../hooks/useCatalog";
import { useApiAction } from "../../../hooks/useApiAction";

// Add / edit / delete products through the API, shared by the Products and Category pages.
export function useProductCrud() {
  const { products, setProducts, categories, reload } = useCatalog();
  const { run, busy } = useApiAction();
  const [editing, setEditing] = useState(null); // null = closed, {} = new, product = edit
  const [deleting, setDeleting] = useState(null);

  const save = (data) =>
    run(
      async () => {
        if (editing?.id !== undefined) {
          const updated = await catalogApi.updateProduct(editing.id, data);
          setProducts((prev) => prev.map((p) => (p.id === updated.id ? updated : p)));
        } else {
          const created = await catalogApi.createProduct(data);
          setProducts((prev) => [created, ...prev]);
        }
        setEditing(null);
        reload(); // refresh category product counts
      },
      ["Product saved", "تم حفظ المنتج"],
    );

  const confirmDelete = () =>
    run(
      async () => {
        await catalogApi.deleteProduct(deleting.id);
        setProducts((prev) => prev.filter((p) => p.id !== deleting.id));
        setDeleting(null);
        reload();
      },
      ["Product deleted", "تم حذف المنتج"],
    );

  return {
    products,
    categories,
    editing,
    deleting,
    busy,
    openAdd: () => setEditing({}),
    openEdit: setEditing,
    closeForm: () => setEditing(null),
    save,
    askDelete: setDeleting,
    cancelDelete: () => setDeleting(null),
    confirmDelete,
  };
}
