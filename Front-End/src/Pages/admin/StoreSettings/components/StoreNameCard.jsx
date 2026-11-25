import { FormField } from "../../../../components/admin/ui";
import SettingsCard from "./SettingsCard";
import styles from "../StoreSettings.module.css";

const MAX_LENGTH = 50;

export default function StoreNameCard({ value, onChange, tr }) {
  return (
    <SettingsCard
      icon="fa-solid fa-store"
      title={tr("Store Name", "اسم المتجر")}
      subtitle={tr("Shown in the header, sidebar and browser tab", "يظهر في الترويسة والقائمة وتبويب المتصفح")}
    >
      <FormField
        name="storeName"
        maxLength={MAX_LENGTH}
        value={value}
        error={value.trim() ? null : tr("Store name is required", "اسم المتجر مطلوب")}
        onChange={(e) => onChange(e.target.value)}
      />
      <span className={styles.counter}>
        {value.length}/{MAX_LENGTH}
      </span>
    </SettingsCard>
  );
}
