import { NavLink } from "react-router";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

import styles from "./PagesList.module.css";

export default function SidebarLink({ page, label, onNavigate }) {
  return (
    <NavLink
      to={page.path}
      end={page.end ?? false}
      className={({ isActive }) => `${styles.pageItem} ${isActive ? styles.active : ""}`}
      onClick={onNavigate}
    >
      <FontAwesomeIcon icon={page.icon} />
      <span>{label}</span>
    </NavLink>
  );
}
