import { useCallback, useState } from "react";

import * as salesApi from "../../../api/sales";
import {
  Card,
  EmptyState,
  IconButton,
  PageHeader,
  SelectInput,
  tableClasses,
  toolbarClass,
} from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useLang } from "../../../hooks/useLang";
import { formatDate } from "../../../utils/format";
import { SaleStatus } from "./shared";
import { useAdminList } from "./salesUtils";
import styles from "./Sales.module.css";

// Customers who asked to be told when an out-of-stock product comes back.
// Restocking a product emails them automatically; "Notify" sends the email now.
export default function WaitingList() {
  const { tr, isAr } = useLang();
  const { run, busy } = useApiAction();
  const [status, setStatus] = useState("waiting");
  const load = useCallback(
    () => salesApi.fetchAdminWaitingList(status === "all" ? "" : status),
    [status],
  );
  const { items, reload } = useAdminList(load);

  const change = (w, next, success) =>
    run(async () => {
      await salesApi.setWaitingStatus(w.id, next);
      await reload();
    }, success);

  return (
    <>
      <PageHeader
        icon="fa-solid fa-bell"
        title={tr("Waiting List", "قائمة الانتظار")}
        subtitle={tr(
          "Customers waiting for out-of-stock products. Restocking emails them automatically.",
          "عملاء ينتظرون منتجات نفدت. إعادة التخزين ترسل لهم بريداً تلقائياً.",
        )}
      />

      <div className={toolbarClass}>
        <SelectInput
          value={status}
          onChange={setStatus}
          aria-label={tr("Status", "الحالة")}
          options={[
            { value: "waiting", label: tr("Waiting", "بالانتظار") },
            { value: "notified", label: tr("Notified", "تم الإبلاغ") },
            { value: "fulfilled", label: tr("Bought", "تم الشراء") },
            { value: "cancelled", label: tr("Cancelled", "ملغى") },
            { value: "all", label: tr("All", "الكل") },
          ]}
        />
      </div>

      {items?.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-bell"
          text={tr("Nobody here.", "لا أحد هنا.")}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Product", "المنتج")}</th>
                  <th>{tr("In stock", "المخزون")}</th>
                  <th>{tr("Customer", "العميل")}</th>
                  <th>{tr("Wants", "الكمية")}</th>
                  <th>{tr("Asked on", "تاريخ الطلب")}</th>
                  <th>{tr("Status", "الحالة")}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items?.map((w) => (
                  <tr key={w.id}>
                    <td>
                      <strong>{isAr ? w.nameAr || w.name : w.name}</strong>
                      {w.color && <div className={styles.muted}>{w.color}</div>}
                    </td>
                    <td className={w.stock > 0 ? "" : styles.owed}>
                      {w.stock}
                    </td>
                    <td>
                      {w.customer}
                      <div className={styles.muted}>{w.customerEmail}</div>
                    </td>
                    <td>{w.quantity}</td>
                    <td>{formatDate(w.requestedAt, isAr)}</td>
                    <td>
                      <SaleStatus status={w.status} />
                    </td>
                    <td>
                      {w.status === "waiting" && (
                        <div
                          className={tableClasses.rowActions}
                          aria-disabled={busy}
                        >
                          <IconButton
                            icon="fa-solid fa-envelope"
                            label={tr(
                              "Notify now (sends the email)",
                              "أبلغ الآن (يرسل البريد)",
                            )}
                            onClick={() =>
                              change(w, "notified", [
                                "Customer notified",
                                "تم إبلاغ العميل",
                              ])
                            }
                          />
                          <IconButton
                            icon="fa-solid fa-ban"
                            danger
                            label={tr("Cancel request", "إلغاء الطلب")}
                            onClick={() =>
                              change(w, "cancelled", [
                                "Request cancelled",
                                "تم إلغاء الطلب",
                              ])
                            }
                          />
                        </div>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}
    </>
  );
}
