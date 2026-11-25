import { useState } from "react";

import * as catalogApi from "../../../api/catalog";
import { useCatalog } from "../../../hooks/useCatalog";
import { useApiAction } from "../../../hooks/useApiAction";

export const isActiveCategory = (category) => category.isActive !== false;

// Category add / edit / delete through the API. The backend cascades renames to the
// category's products and hides its products on delete, so we reload the catalogue after.
export function useCategoryCrud() {
  const { categories, products, reload } = useCatalog();
  const { run, busy } = useApiAction();
  const [editing, setEditing] = useState(null); // null = closed, {} = new
  const [deleting, setDeleting] = useState(null);

  const countFor = (category) => products.filter((p) => p.category === category.name).length;

  const save = (data) =>
    run(
      async () => {
        if (editing?.id === undefined) await catalogApi.createCategory(data);
        else await catalogApi.updateCategory(editing.id, data);
        await reload();
        setEditing(null);
      },
      ["Category saved", "تم حفظ التصنيف"],
    );

  const confirmDelete = () =>
    run(
      async () => {
        await catalogApi.deleteCategory(deleting.id);
        await reload();
        setDeleting(null);
      },
      ["Category deleted", "تم حذف التصنيف"],
    );

  return {
    categories,
    countFor,
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
