import { useState } from "react";

import { uploadSettingsImage } from "../../../../api/catalog";
import { Button, FormField, formGridClass } from "../../../../components/admin/ui";
import {
  DEFAULT_FEATURES,
  DEFAULT_FOOTER,
  DEFAULT_HERO,
  DEFAULT_PRODUCTS_BANNER,
  SOCIAL_NETWORKS,
} from "../../../../data/storeContent";
import { useToast } from "../../../../hooks/useToast";
import SettingsCard from "./SettingsCard";
import styles from "../StoreSettings.module.css";

const MAX_FEATURES = 8;

// Page content of this store: banners, feature tiles, footer. Empty fields
// keep the default text (shown as the placeholder), so an admin only fills in
// what they want to change. Saved with the rest of the settings.
export default function ContentCard({ content, onChange, tr }) {
  const c = content ?? {};
  const set = (section, value) => onChange({ ...c, [section]: value });

  return (
    <SettingsCard
      icon="fa-solid fa-pen-ruler"
      title={tr("Page content", "محتوى الصفحات")}
      subtitle={tr(
        "Banners, feature tiles and footer of your store. Empty fields keep the default text.",
        "اللافتات والمزايا وتذييل متجرك. الحقول الفارغة تبقي النص الافتراضي.",
      )}
    >
      <BannerEditor
        title={tr("Home page banner", "لافتة الصفحة الرئيسية")}
        value={c.hero}
        defaults={DEFAULT_HERO}
        onChange={(v) => set("hero", v)}
        tr={tr}
      />
      <BannerEditor
        title={tr("Products page banner", "لافتة صفحة المنتجات")}
        value={c.productsBanner}
        defaults={DEFAULT_PRODUCTS_BANNER}
        onChange={(v) => set("productsBanner", v)}
        tr={tr}
      />
      <FeaturesEditor value={c.features} onChange={(v) => set("features", v)} tr={tr} />
      <FooterEditor value={c.footer} onChange={(v) => set("footer", v)} tr={tr} />
    </SettingsCard>
  );
}

function Section({ title, children }) {
  return (
    <fieldset className={styles.contentSection}>
      <legend>{title}</legend>
      {children}
    </fieldset>
  );
}

// Two inputs (English / Arabic) for one text.
function Pair({ label, value, defaults, keyEn, keyAr, onChange, tr }) {
  return (
    <div className={formGridClass}>
      <FormField label={label} name={keyEn} value={value[keyEn] ?? ""} placeholder={defaults[keyEn]}
        onChange={(e) => onChange({ ...value, [keyEn]: e.target.value })} />
      <FormField label={`${label} (${tr("Arabic", "عربي")})`} name={keyAr} dir="rtl" value={value[keyAr] ?? ""}
        placeholder={defaults[keyAr]} onChange={(e) => onChange({ ...value, [keyAr]: e.target.value })} />
    </div>
  );
}

function BannerEditor({ title, value, defaults, onChange, tr }) {
  const v = value ?? {};
  const toast = useToast();
  const [uploading, setUploading] = useState(false);

  const pick = async (file) => {
    if (!file) return;
    setUploading(true);
    try {
      onChange({ ...v, image: await uploadSettingsImage(file) });
    } catch (err) {
      toast.error(err.message, err.message);
    } finally {
      setUploading(false);
    }
  };

  return (
    <Section title={title}>
      <Pair label={tr("Title", "العنوان")} value={v} defaults={defaults} keyEn="title" keyAr="titleAr" onChange={onChange} tr={tr} />
      <Pair label={tr("Text", "النص")} value={v} defaults={defaults} keyEn="subtitle" keyAr="subtitleAr" onChange={onChange} tr={tr} />
      <Pair label={tr("Highlighted text", "النص المميز")} value={v} defaults={defaults} keyEn="highlight" keyAr="highlightAr" onChange={onChange} tr={tr} />
      <Pair label={tr("Button", "الزر")} value={v} defaults={defaults} keyEn="buttonText" keyAr="buttonTextAr" onChange={onChange} tr={tr} />
      <div className={styles.imageRow}>
        <img src={v.image || defaults.image} alt="" />
        <label className={styles.fileButton}>
          <i className="fa-solid fa-upload" /> {uploading ? tr("Uploading…", "جارٍ الرفع…") : tr("Upload image", "رفع صورة")}
          <input type="file" accept="image/jpeg,image/png,image/webp" hidden disabled={uploading}
            onChange={(e) => pick(e.target.files[0])} />
        </label>
        {v.image && (
          <Button variant="ghost" onClick={() => onChange({ ...v, image: "" })}>
            {tr("Use the default image", "استخدام الصورة الافتراضية")}
          </Button>
        )}
      </div>
    </Section>
  );
}

function FeaturesEditor({ value, onChange, tr }) {
  // Editing starts from the defaults, so the admin sees what is shown now.
  const list = value?.length ? value : DEFAULT_FEATURES;
  const update = (i, key, v) => onChange(list.map((f, j) => (j === i ? { ...f, [key]: v } : f)));

  return (
    <Section title={tr("Feature tiles", "المزايا")}>
      {list.map((f, i) => (
        <div key={i} className={styles.featureRow}>
          <FormField label={tr("Icon", "الأيقونة")} name={`icon${i}`} value={f.icon} hint="fa-solid fa-truck"
            onChange={(e) => update(i, "icon", e.target.value)} />
          <FormField label={tr("Name", "الاسم")} name={`name${i}`} value={f.name} onChange={(e) => update(i, "name", e.target.value)} />
          <FormField label={tr("Name (Arabic)", "الاسم (عربي)")} name={`nameAr${i}`} dir="rtl" value={f.nameAr}
            onChange={(e) => update(i, "nameAr", e.target.value)} />
          <FormField label={tr("Line", "السطر")} name={`comment${i}`} value={f.comment} onChange={(e) => update(i, "comment", e.target.value)} />
          <FormField label={tr("Line (Arabic)", "السطر (عربي)")} name={`commentAr${i}`} dir="rtl" value={f.commentAr}
            onChange={(e) => update(i, "commentAr", e.target.value)} />
          <Button variant="ghost" icon="fa-solid fa-trash" onClick={() => onChange(list.filter((_, j) => j !== i))}>
            {tr("Remove", "إزالة")}
          </Button>
        </div>
      ))}
      <div className={styles.rowButtons}>
        <Button variant="outline" icon="fa-solid fa-plus" disabled={list.length >= MAX_FEATURES}
          onClick={() => onChange([...list, { icon: "fa-solid fa-star", name: "", nameAr: "", comment: "", commentAr: "" }])}>
          {tr("Add a tile", "إضافة ميزة")}
        </Button>
        {value?.length > 0 && (
          <Button variant="ghost" onClick={() => onChange([])}>
            {tr("Back to the default tiles", "العودة إلى المزايا الافتراضية")}
          </Button>
        )}
      </div>
    </Section>
  );
}

function FooterEditor({ value, onChange, tr }) {
  const v = value ?? {};
  const social = v.social ?? {};
  return (
    <Section title={tr("Footer", "التذييل")}>
      <Pair label={tr("Tagline", "الشعار النصي")} value={v} defaults={{ tagline: "Better Choose Smarter Shopping", taglineAr: "اختر بشكل أفضل تسوّق بذكاء أكبر" }}
        keyEn="tagline" keyAr="taglineAr" onChange={onChange} tr={tr} />
      <Pair label={tr("Copyright", "حقوق النشر")} value={v} defaults={DEFAULT_FOOTER} keyEn="copyright" keyAr="copyrightAr" onChange={onChange} tr={tr} />
      <div className={formGridClass}>
        {Object.entries(SOCIAL_NETWORKS).map(([network, icon]) => (
          <FormField key={network} label={<><i className={icon} /> {network}</>} name={`social-${network}`}
            value={social[network] ?? ""} placeholder="https://"
            onChange={(e) => onChange({ ...v, social: { ...social, [network]: e.target.value.trim() } })} />
        ))}
      </div>
    </Section>
  );
}
