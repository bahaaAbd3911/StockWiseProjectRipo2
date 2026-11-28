import { useBrand } from "../../../hooks/useBrand";
import { useLang } from "../../../hooks/useLang";
import { LEGAL_UPDATED, PRIVACY, TERMS } from "../../../data/legalPages";
import { formatDate } from "../../../utils/format";
import UserSettings from "../../../components/common/UserSettings/UserSettings";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import Footer from "../../../components/common/Footer/Footer";
import styles from "./LegalPage.module.css";

const PAGES = { privacy: PRIVACY, terms: TERMS };

// /client/privacy and /client/terms: texts from data/legalPages.js, with the
// store's name filled in.
export default function LegalPage({ page }) {
  const { isAr } = useLang();
  const { name } = useBrand();
  const doc = PAGES[page];
  const t = ([en, ar]) => (isAr ? ar : en).replaceAll("{store}", name);

  return (
    <>
      <UserSettings />
      <SearchBar value={1} />
      <article className={styles.page}>
        <h1>{t(doc.title)}</h1>
        <p className={styles.intro}>{t(doc.intro)}</p>
        {doc.sections.map((s) => (
          <section key={s.title[0]}>
            <h2>{t(s.title)}</h2>
            <ul>
              {s.items.map((item) => (
                <li key={item[0]}>{t(item)}</li>
              ))}
            </ul>
          </section>
        ))}
        <p className={styles.updated}>
          {isAr ? "آخر تحديث" : "Last updated"}:{" "}
          {formatDate(LEGAL_UPDATED, isAr)}
        </p>
      </article>
      <Footer />
    </>
  );
}
