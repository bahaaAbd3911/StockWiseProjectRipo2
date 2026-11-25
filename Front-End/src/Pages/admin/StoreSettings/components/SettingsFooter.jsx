import { Button, Card } from "../../../../components/admin/ui";
import styles from "../StoreSettings.module.css";

export default function SettingsFooter({ canSave, onReset, onSave, tr }) {
  return (
    <Card className={styles.footer}>
      <div className={styles.footerText}>
        <i className="fa-solid fa-clock-rotate-left" />
        <div>
          <h4>{tr("Reset to Default Settings", "استعادة الإعدادات الافتراضية")}</h4>
          <p>{tr("Restore all settings to the default theme, then save", "أعد كل الإعدادات إلى الافتراضي ثم احفظ")}</p>
        </div>
      </div>

      <div className={styles.footerActions}>
        <Button variant="ghost" icon="fa-solid fa-rotate-left" onClick={onReset}>{tr("Reset to Default", "استعادة الافتراضي")}</Button>
        <Button icon="fa-solid fa-circle-check" disabled={!canSave} onClick={onSave}>{tr("Save Changes", "حفظ التغييرات")}</Button>
      </div>
    </Card>
  );
}
