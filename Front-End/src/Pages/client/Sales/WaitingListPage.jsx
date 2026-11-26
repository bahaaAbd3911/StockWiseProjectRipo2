import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router";

import { useLang } from "../../../hooks/useLang";
import { useToast } from "../../../hooks/useToast";
import { cancelWaiting, fetchMyWaitingList } from "../../../api/sales";
import { formatDate } from "../../../utils/format";
import Header from "../../../components/client/Header/Header";
import favoriteImg from "../../../assets/images/Favorites/favoriteLogo.png";
import styles from "./Sales.module.css";

const STATUS = {
  waiting: ["Waiting", "بالانتظار"],
  notified: ["Back in stock", "متوفر الآن"],
  fulfilled: ["Bought", "تم الشراء"],
  cancelled: ["Cancelled", "ملغى"],
};

// "Waiting List" tab: products the customer asked to be told about when back in stock.
export default function WaitingListPage() {
  const { tr, isAr } = useLang();
  const toast = useToast();
  const navigate = useNavigate();
  const [items, setItems] = useState(null);

  const load = useCallback(
    () =>
      fetchMyWaitingList()
        .then(setItems)
        .catch((err) => {
          setItems([]);
          toast.error(err.message, err.message);
        }),
    [toast],
  );

  useEffect(() => {
    load();
  }, [load]);

  const cancel = async (id) => {
    try {
      await cancelWaiting(id);
      toast.success(
        "Removed from your waiting list",
        "تمت الإزالة من قائمة الانتظار",
      );
      load();
    } catch (err) {
      toast.error(err.message, err.message);
    }
  };

  return (
    <div className={styles.page}>
      <Header
        icon={"fa-solid fa-bell"}
        title={"Waiting List"}
        arbTitle={"قائمة الانتظار"}
        subTitle={"We email you when these products are back in stock."}
        arbSubTitle={"سنرسل لك بريداً عندما تعود هذه المنتجات إلى المخزون."}
        img={favoriteImg}
      />

      {items === null && (
        <p className={styles.muted}>{tr("Loading…", "جارٍ التحميل…")}</p>
      )}
      {items?.length === 0 && (
        <p className={styles.note}>
          {tr(
            'Nothing here. On an out-of-stock product, press "Notify me".',
            'لا شيء هنا. اضغط "أعلمني" في صفحة أي منتج نفد من المخزون.',
          )}
        </p>
      )}

      <div className={styles.list}>
        {items?.map((w) => {
          const [en, ar] = STATUS[w.status] ?? [w.status, w.status];
          return (
            <div className={styles.row} key={w.id}>
              <div>
                <strong
                  onClick={() =>
                    navigate(`/client/ProductDetails/${w.productId}`)
                  }
                >
                  {isAr ? w.nameAr || w.name : w.name}
                </strong>
                <div className={styles.muted}>
                  {tr("Quantity", "الكمية")}: {w.quantity}
                  {w.color && ` · ${w.color}`} ·{" "}
                  {formatDate(w.requestedAt, isAr)}
                </div>
              </div>
              <span
                className={`${styles.badge} ${w.status === "notified" ? styles.badgeOn : ""}`}
              >
                {tr(en, ar)}
              </span>
              {(w.status === "waiting" || w.status === "notified") && (
                <button
                  type="button"
                  className={`${styles.btn} ${styles.ghost}`}
                  onClick={() => cancel(w.id)}
                >
                  {tr("Remove", "إزالة")}
                </button>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
