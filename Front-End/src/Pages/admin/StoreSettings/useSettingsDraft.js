import { useContext, useState } from "react";

import { StoreContext } from "../../../context/StoreContext";
import { DEFAULT_STORE_SETTINGS } from "../../../data/storeSettings";
import { useToast } from "../../../hooks/useToast";
import { isHexColor } from "../../../utils/validators";

// Edits a copy of the store settings; nothing changes in the store until save().
export function useSettingsDraft() {
  const { storeSettings, saveStoreSettings } = useContext(StoreContext);
  const toast = useToast();
  const [draft, setDraft] = useState(storeSettings);
  const [saving, setSaving] = useState(false);

  const setField = (key, value) => setDraft((prev) => ({ ...prev, [key]: value }));

  const setColor = (mode, key, value) =>
    setDraft((prev) => ({
      ...prev,
      colors: { ...prev.colors, [mode]: { ...prev.colors[mode], [key]: value } },
    }));

  const allColorsValid = Object.values(draft.colors).every((set) => Object.values(set).every(isHexColor));
  const isValid = draft.storeName.trim().length > 0 && allColorsValid;

  const save = async () => {
    setSaving(true);
    try {
      const saved = await saveStoreSettings({ ...draft, storeName: draft.storeName.trim() });
      setDraft(saved);
      toast.success("Store settings saved", "تم حفظ إعدادات المتجر");
    } catch (err) {
      toast.error(err.message, err.message);
    } finally {
      setSaving(false);
    }
  };

  return {
    draft,
    setField,
    setColor,
    isValid,
    saving,
    isDirty: JSON.stringify(draft) !== JSON.stringify(storeSettings),
    resetToDefault: () => setDraft(DEFAULT_STORE_SETTINGS),
    save,
  };
}
