import { splitBrandName } from "../../../../hooks/useBrand";
import SettingsCard from "./SettingsCard";
import styles from "./LivePreview.module.css";

// A mini storefront painted with the draft colors of the selected mode.
export default function LivePreview({ settings, mode, tr }) {
  const c = settings.colors[mode];
  const [first, rest] = splitBrandName(settings.storeName || " ");
  const vars = {
    "--p-text": c.text,
    "--p-border": c.border,
    "--p-bg": c.background,
    "--p-button": c.button,
    "--p-button-hover": c.buttonHover,
    "--p-accent": c.accent,
    "--p-surface": c.surface,
    "--p-muted": c.mutedText,
    "--p-danger": c.danger,
  };

  return (
    <SettingsCard icon="fa-solid fa-eye" title={tr("Live Preview", "معاينة مباشرة")} subtitle={tr("See how your changes look", "شاهد شكل تغييراتك")}>
      <div className={styles.store} style={vars}>
        <div className={styles.header}>
          <span className={styles.brand}>
            {settings.logo ? <img src={settings.logo} alt="" style={{ height: 22 }} /> : <i className="fa-solid fa-basket-shopping" />}
            {first}
            <span>{rest}</span>
          </span>
          <span className={styles.search}>{tr("Search for products…", "ابحث عن منتج…")}</span>
          <i className="fa-solid fa-cart-shopping" />
        </div>

        <div className={styles.nav}>
          <span className={styles.navActive}>{tr("Home", "الرئيسية")}</span>
          <span>{tr("Categories", "التصنيفات")}</span>
          <span>{tr("Products", "المنتجات")}</span>
        </div>

        <div className={styles.hero}>
          <h4>
            {tr("Your One-Stop ", "متجرك الشامل ")}
            <span>{tr("Online Store", "على الإنترنت")}</span>
          </h4>
          <p>{tr("Shop smarter. Live better.", "تسوّق بذكاء. عِش أفضل.")}</p>
          <button type="button" className={styles.button}>{tr("Shop Now", "تسوّق الآن")} →</button>
        </div>

        <div className={styles.samples}>
          <button type="button" className={styles.button}>{tr("Button", "زر")}</button>
          <span className={styles.link}>{tr("Link", "رابط")} →</span>
          <div className={styles.sampleCard}>
            <strong>{tr("Sample Card", "بطاقة تجريبية")}</strong>
            <p>{tr("This is how your text looks.", "هكذا يبدو النص.")}</p>
          </div>
        </div>
      </div>
    </SettingsCard>
  );
}
