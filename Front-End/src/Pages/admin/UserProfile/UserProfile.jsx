import { useState } from "react";
import { Link, Navigate, useNavigate, useParams } from "react-router";

import { useLang } from "../../../hooks/useLang";
import { useOrderStatusUpdate } from "../../../hooks/useOrderStatusUpdate";
import { EmptyState } from "../../../components/admin/ui";
import OrdersTable from "../../../components/admin/orders/OrdersTable";
import OrderDetailsModal from "../../../components/admin/orders/OrderDetailsModal";
import { useUserCrud } from "../Users/useUserCrud";
import UserModals from "../Users/components/UserModals";
import { useUserProfileData } from "./useUserProfileData";
import ProfileCard from "./components/ProfileCard";
import ProfileStats from "./components/ProfileStats";
import UserFavorites from "./components/UserFavorites";
import styles from "./UserProfile.module.css";

export default function UserProfile() {
  const { userId } = useParams();
  const navigate = useNavigate();
  const { tr, isAr } = useLang();
  const [openOrderId, setOpenOrderId] = useState(null);
  const { user, userOrders, favorites, status, reload } = useUserProfileData(userId);
  const crud = useUserCrud({ onSaved: reload, onDeleted: () => navigate("/admin/users", { replace: true }) });
  const { changeStatus } = useOrderStatusUpdate(reload);

  if (status === "missing") return <Navigate to="/admin/users" replace />;

  return (
    <>
      <Link to="/admin/users" className={styles.back}>
        <i className={`fa-solid ${isAr ? "fa-arrow-right" : "fa-arrow-left"}`} />
        {tr("Back to users", "العودة إلى المستخدمين")}
      </Link>

      {status === "loading" && <EmptyState icon="fa-solid fa-spinner fa-spin" text={tr("Loading…", "جارٍ التحميل…")} />}

      {user && (
        <>
          <ProfileCard user={user} isAr={isAr} tr={tr} onEdit={() => crud.openEdit(user)} onDelete={() => crud.askDelete(user)} />
          <ProfileStats orders={userOrders} favoritesCount={favorites.length} tr={tr} />
          <UserFavorites products={favorites} isAr={isAr} tr={tr} />
          <h2 className={styles.sectionTitle}>{tr("Order History", "سجل الطلبات")}</h2>
          <OrdersTable orders={userOrders} onOpen={setOpenOrderId} />
        </>
      )}

      <UserModals crud={crud} isAr={isAr} tr={tr} />
      {openOrderId && (
        <OrderDetailsModal
          order={userOrders.find((o) => o.id === openOrderId)}
          onChangeStatus={changeStatus}
          onClose={() => setOpenOrderId(null)}
        />
      )}
    </>
  );
}
