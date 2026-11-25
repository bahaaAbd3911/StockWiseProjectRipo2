import { COLOR_TOKENS } from "../../../../data/storeSettings";
import SettingsCard from "./SettingsCard";
import ColorField from "./ColorField";
import styles from "../StoreSettings.module.css";

export default function ColorsCard({ colors, mode, onModeChange, onColor, tr }) {
  const modes = [
    ["light", "fa-solid fa-sun", tr("Light Mode", "الوضع الفاتح")],
    ["dark", "fa-solid fa-moon", tr("Dark Mode", "الوضع الداكن")],
  ];

  return (
    <SettingsCard
      icon="fa-solid fa-palette"
      title={tr("Colors & Style", "الألوان والنمط")}
      subtitle={tr(
        "These colors are used everywhere in the store and admin panel. Each mode has its own set.",
        "تُستخدم هذه الألوان في كل المتجر ولوحة التحكم. لكل وضع ألوانه الخاصة.",
      )}
    >
      <div className={styles.modeTabs} role="tablist">
        {modes.map(([key, icon, label]) => (
          <button
            key={key}
            role="tab"
            aria-selected={mode === key}
            className={`${styles.modeTab} ${mode === key ? styles.modeTabActive : ""}`}
            onClick={() => onModeChange(key)}
          >
            <i className={icon} /> {label}
          </button>
        ))}
      </div>

      <div className={styles.colorsGrid}>
        {COLOR_TOKENS.map((token) => (
          <ColorField
            key={token.key}
            label={tr(token.en, token.ar)}
            hint={tr(token.hintEn, token.hintAr)}
            value={colors[mode][token.key]}
            onChange={(value) => onColor(mode, token.key, value)}
          />
        ))}
      </div>
    </SettingsCard>
  );
}
