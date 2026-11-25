import { useCallback, useEffect, useState } from "react";

import * as adminApi from "../../../api/admin";
import { useApiAction } from "../../../hooks/useApiAction";

// Loads users from the API and handles add / edit / delete.
// `onChanged` lets the profile page refresh after an edit or delete.
export function useUserCrud({ onSaved, onDeleted } = {}) {
  const { run, busy } = useApiAction();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(null); // null = closed, {} = new
  const [deleting, setDeleting] = useState(null);

  const reload = useCallback(async () => {
    try {
      setUsers(await adminApi.fetchUsers());
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    reload().catch(() => {});
  }, [reload]);

  const save = (form) =>
    run(
      async () => {
        const saved =
          editing?.userID === undefined
            ? await adminApi.createUser(form)
            : await adminApi.updateUser(editing.userID, form);
        setEditing(null);
        await reload();
        onSaved?.(saved);
      },
      ["User saved", "تم حفظ المستخدم"],
    );

  const confirmDelete = () =>
    run(
      async () => {
        await adminApi.deleteUser(deleting.userID);
        setDeleting(null);
        await reload();
        onDeleted?.();
      },
      ["User deleted", "تم حذف المستخدم"],
    );

  return {
    users,
    loading,
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
