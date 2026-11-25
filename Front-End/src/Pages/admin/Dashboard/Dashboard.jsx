import { useContext, useState } from "react";

import { AuthContext } from "../../../context/AuthContext";
import { formatDate } from "../../../utils/format";
import { translations, STAT_ICONS } from "./dashboardData";
import { useDashboardData } from "./useDashboardData";

import StatCard from "./components/StatCard";
import SalesChart from "./components/SalesChart";
import ActivityPanel from "./components/ActivityPanel";
import OrdersPanel from "./components/OrdersPanel";
import UsersPanel from "./components/UsersPanel";
import ProductsPanel from "./components/ProductsPanel";

import styles from "./Dashboard.module.css";

function Dashboard() {
  const { dark, lang, user } = useContext(AuthContext);
  const [period, setPeriod] = useState("weekly");
  const data = useDashboardData();

  const isArabic = lang === "arb";
  const t = translations[isArabic ? "arb" : "en"];

  return (
    <main
      className={`${styles.dashboard} ${dark ? styles.dark : styles.light}`}
      dir={isArabic ? "rtl" : "ltr"}
    >
      <header className={styles.pageHeader}>
        <div>
          <h1>{t.dashboard}</h1>
          <h2>{`${t.welcomeName}, ${user?.firstName ?? ""}!`}</h2>
          <p>{t.overview}</p>
        </div>

        <span className={styles.todayDate}>{formatDate(new Date(), isArabic)}</span>
      </header>

      <section className={styles.statsGrid}>
        {data.stats.map((item) => (
          <StatCard key={item.key} item={{ ...item, icon: STAT_ICONS[item.key] }} t={t} />
        ))}
      </section>

      <section className={styles.middleGrid}>
        <SalesChart t={t} period={period} setPeriod={setPeriod} orders={data.orders} isAr={isArabic} />
        <ActivityPanel t={t} activities={data.activities} />
      </section>

      <OrdersPanel t={t} orders={data.recentOrders} isAr={isArabic} />

      <section className={styles.bottomGrid}>
        <UsersPanel t={t} users={data.recentUsers} isAr={isArabic} />
        <ProductsPanel t={t} rows={data.topSellers} isAr={isArabic} />
      </section>
    </main>
  );
}

export default Dashboard;
