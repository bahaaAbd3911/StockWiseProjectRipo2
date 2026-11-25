import { useState } from "react";

import { useLang } from "../../../hooks/useLang";
import { PageHeader } from "../../../components/admin/ui";
import { useSettingsDraft } from "./useSettingsDraft";
import LogoCard from "./components/LogoCard";
import StoreNameCard from "./components/StoreNameCard";
import ColorsCard from "./components/ColorsCard";
import ContentCard from "./components/ContentCard";
import LivePreview from "./components/LivePreview";
import SettingsFooter from "./components/SettingsFooter";
import styles from "./StoreSettings.module.css";

export default function StoreSettings() {
  const { tr, dark } = useLang();
  const settings = useSettingsDraft();
  const { draft } = settings;
  // Start on the color set that matches the current theme.
  const [mode, setMode] = useState(dark ? "dark" : "light");

  return (
    <>
      <PageHeader
        icon="fa-solid fa-gear"
        title={tr("Store Settings", "إعدادات المتجر")}
        subtitle={tr("Customize your store's name, logo and colors.", "خصّص اسم متجرك وشعاره وألوانه.")}
      />

      <div className={styles.layout}>
        <div className={styles.column}>
          <div className={styles.twoCols}>
            <LogoCard logo={draft.logo} storeName={draft.storeName} onLogo={(v) => settings.setField("logo", v)} tr={tr} />
            <StoreNameCard value={draft.storeName} onChange={(v) => settings.setField("storeName", v)} tr={tr} />
          </div>
          <ColorsCard colors={draft.colors} mode={mode} onModeChange={setMode} onColor={settings.setColor} tr={tr} />
          <ContentCard content={draft.content} onChange={(v) => settings.setField("content", v)} tr={tr} />
        </div>

        <LivePreview settings={draft} mode={mode} tr={tr} />
      </div>

      <SettingsFooter
        canSave={settings.isDirty && settings.isValid && !settings.saving}
        onReset={settings.resetToDefault}
        onSave={settings.save}
        tr={tr}
      />
    </>
  );
}
