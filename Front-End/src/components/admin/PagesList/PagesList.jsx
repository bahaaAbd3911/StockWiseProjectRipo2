import { useState, useContext } from "react";
import { useNavigate } from "react-router";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faArrowRightFromBracket, faBars, faXmark } from "@fortawesome/free-solid-svg-icons";

import { AuthContext } from "../../../context/AuthContext";
import { useBrand } from "../../../hooks/useBrand";
import BrandMark from "../../common/BrandLogo/BrandMark";
import SidebarLink from "./SidebarLink";
import { translations, mainPages, websitePages } from "./pagesConfig";

import styles from "./PagesList.module.css";

function PagesList() {
  const { dark, lang, signOut } = useContext(AuthContext);
  const { first, rest } = useBrand();
  const navigate = useNavigate();
  const [isOpen, setIsOpen] = useState(false);

  const isArabic = lang === "arb";
  const t = translations[lang] || translations.en;
  const close = () => setIsOpen(false);

  const logout = () => {
    signOut();
    navigate("/signIn");
  };

  const renderLink = (page) => (
    <SidebarLink key={page.path} page={page} label={t[page.key]} onNavigate={close} />
  );

  return (
    <div
      className={`${styles.sidebarWrapper} ${dark ? styles.dark : styles.light} ${isArabic ? styles.arabic : ""}`}
      dir={isArabic ? "rtl" : "ltr"}
    >
      <button
        className={styles.menuButton}
        onClick={() => setIsOpen(!isOpen)}
        aria-label={isOpen ? "Close navigation" : "Open navigation"}
      >
        <FontAwesomeIcon icon={isOpen ? faXmark : faBars} />
      </button>

      {isOpen && <button className={styles.overlay} onClick={close} aria-label="Close navigation" />}

      <aside className={`${styles.pagesList} ${isOpen ? styles.open : ""}`}>
        <div className={styles.logo}>
          <BrandMark className={styles.logoIcon} size={43} />
          <h2>
            {first}
            <span>{rest}</span>
          </h2>
          <button className={styles.closeButton} onClick={close} aria-label="Close navigation">
            <FontAwesomeIcon icon={faXmark} />
          </button>
        </div>

        <nav className={styles.navigation}>
          <div className={styles.pagesGroup}>{mainPages.map(renderLink)}</div>
          <div className={styles.websiteGroup}>
            <p className={styles.groupTitle}>{t.website}</p>
            {websitePages.map(renderLink)}
          </div>
        </nav>

        <button className={styles.logout} onClick={logout}>
          <FontAwesomeIcon icon={faArrowRightFromBracket} />
          <span>{t.logout}</span>
        </button>
      </aside>
    </div>
  );
}

export default PagesList;
