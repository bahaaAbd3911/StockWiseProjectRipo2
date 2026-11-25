import { IconButton } from "../../../../components/admin/ui";
import { splitBrandName } from "../../../../hooks/useBrand";
import SettingsCard from "./SettingsCard";
import LogoDropZone from "./LogoDropZone";
import styles from "../StoreSettings.module.css";

export default function LogoCard({ logo, storeName, onLogo, tr }) {
  const [first, rest] = splitBrandName(storeName || " ");

  return (
    <SettingsCard
      icon="fa-regular fa-image"
      title={tr("Store Logo", "شعار المتجر")}
      subtitle={tr("Upload your logo or keep the default one", "ارفع شعارك أو استخدم الشعار الافتراضي")}
    >
      <div className={styles.logoRow}>
        <LogoDropZone onLogo={onLogo} tr={tr} />

        <div className={styles.logoPreview}>
          {logo ? <img src={logo} alt="" style={{ maxHeight: 70, maxWidth: "80%" }} /> : <i className="fa-solid fa-basket-shopping" />}
          <span>
            {first}
            <span>{rest}</span>
          </span>
          {logo && (
            <span className={styles.removeLogo}>
              <IconButton icon="fa-solid fa-trash" danger label={tr("Remove logo", "إزالة الشعار")} onClick={() => onLogo("")} />
            </span>
          )}
        </div>
      </div>
    </SettingsCard>
  );
}
