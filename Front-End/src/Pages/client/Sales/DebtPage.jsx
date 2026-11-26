import { useEffect, useState } from "react";

import { useLang } from "../../../hooks/useLang";
import { fetchMyDebt } from "../../../api/sales";
import { formatDate, formatMoney } from "../../../utils/format";
import Header from "../../../components/client/Header/Header";
import cartImg from "../../../assets/images/ShoppingCart/cartLogo.png";
import styles from "./Sales.module.css";

const TYPE_LABELS = {
  charge: ["Order on credit", "طلب بالدين"],
  payment: ["Payment", "دفعة"],
  reversal: ["Order cancelled", "إلغاء طلب"],
  adjustment: ["Adjustment", "تعديل"],
};

// "My Balance" tab: what the customer owes the store, their credit limit and history.
export default function DebtPage() {
  const { tr, isAr } = useLang();
  const [state, setState] = useState({ loading: true, data: null, error: "" });

  useEffect(() => {
    let current = true;
    fetchMyDebt()
      .then((data) => current && setState({ loading: false, data, error: "" }))
      .catch(
        (err) =>
          current &&
          setState({ loading: false, data: null, error: err.message }),
      );
    return () => {
      current = false;
    };
  }, []);

  const { loading, data, error } = state;
  const a = data?.account;

  return (
    <div className={styles.page}>
      <Header
        icon={"fa-solid fa-wallet"}
        title={"My Balance"}
        arbTitle={"رصيدي"}
        subTitle={"What you owe the store and your pay-later history."}
        arbSubTitle={"المبلغ المستحق عليك وسجل الدفع لاحقاً."}
        img={cartImg}
      />

      {loading && (
        <p className={styles.muted}>{tr("Loading…", "جارٍ التحميل…")}</p>
      )}
      {error && <p className={styles.owed}>{error}</p>}

      {a && (
        <>
          <div className={styles.cards}>
            <div className={styles.stat}>
              <small>{tr("You owe", "المبلغ المستحق")}</small>
              <strong className={a.balance > 0 ? styles.owed : ""}>
                {formatMoney(a.balance)}
              </strong>
            </div>
            <div className={styles.stat}>
              <small>{tr("Credit limit", "حد الائتمان")}</small>
              <strong>{formatMoney(a.creditLimit)}</strong>
            </div>
            <div className={styles.stat}>
              <small>
                {tr("Available to pay later", "المتاح للدفع لاحقاً")}
              </small>
              <strong className={styles.good}>
                {formatMoney(a.availableCredit)}
              </strong>
            </div>
          </div>

          {!a.payLaterAllowed && (
            <p className={styles.note}>
              {tr(
                "Pay later is not available on your account. Ask the store if you would like it.",
                "الدفع لاحقاً غير متاح لحسابك. تواصل مع المتجر إذا رغبت بتفعيله.",
              )}
            </p>
          )}

          {data.transactions.length === 0 ? (
            <p className={styles.note}>
              {tr("No transactions yet.", "لا توجد حركات بعد.")}
            </p>
          ) : (
            <div className={styles.tableWrap}>
              <table className={styles.table}>
                <thead>
                  <tr>
                    <th>{tr("Date", "التاريخ")}</th>
                    <th>{tr("Type", "النوع")}</th>
                    <th>{tr("Amount", "المبلغ")}</th>
                    <th>{tr("Balance after", "الرصيد بعدها")}</th>
                    <th>{tr("Note", "ملاحظة")}</th>
                  </tr>
                </thead>
                <tbody>
                  {data.transactions.map((t) => {
                    const [en, ar] = TYPE_LABELS[t.type] ?? [t.type, t.type];
                    return (
                      <tr key={t.id}>
                        <td>{formatDate(t.createdAt, isAr, true)}</td>
                        <td>
                          {tr(en, ar)}
                          {t.orderId ? ` #${t.orderId}` : ""}
                        </td>
                        <td
                          className={t.amount > 0 ? styles.owed : styles.good}
                        >
                          {formatMoney(t.amount)}
                        </td>
                        <td>{formatMoney(t.balanceAfter)}</td>
                        <td className={styles.muted}>{t.note}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </>
      )}
    </div>
  );
}
