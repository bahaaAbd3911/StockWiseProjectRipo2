import { Outlet } from "react-router";

import UserSettings from "../../components/common/UserSettings/UserSettings";
import PagesList from "../../components/admin/PagesList/PagesList";
import { useLang } from "../../hooks/useLang";
import AdminGuard from "./AdminGuard";

import "./adminTheme.css";
import styles from "./AdminLayout.module.css";

// Sidebar + top settings bar; the active admin page renders in <Outlet />.
function AdminLayout() {
  const { dark, dir } = useLang();

  return (
    <AdminGuard>
      <div className={`adminTheme ${dark ? "adminThemeDark" : ""}`} dir={dir}>
        <UserSettings />
        <div className={styles.mainContainerClass}>
          <PagesList />
          <main className={styles.content}>
            <Outlet />
          </main>
        </div>
      </div>
    </AdminGuard>
  );
}

export default AdminLayout;
