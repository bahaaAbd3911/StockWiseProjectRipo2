import { useCallback, useEffect, useState } from "react";

import * as salesApi from "../../../api/sales";
import { fetchUsers } from "../../../api/admin";
import {
  Button,
  Card,
  EmptyState,
  FormField,
  Modal,
  PageHeader,
  SelectInput,
  StatCard,
  StatsGrid,
  tableClasses,
  toolbarClass,
} from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useLang } from "../../../hooks/useLang";
import { formatDate, formatMoney } from "../../../utils/format";
import { useAdminList } from "./salesUtils";
import styles from "./Sales.module.css";

const TYPE_LABELS = {
  charge: ["Order on credit", "طلب بالدين"],
  payment: ["Payment", "دفعة"],
  reversal: ["Order cancelled", "إلغاء طلب"],
  adjustment: ["Adjustment", "تعديل"],
};

// Pay later: which customers may buy on credit, how much they owe, and
// recording the money they bring in.
export default function Debts() {
  const { tr } = useLang();
  const [outstanding, setOutstanding] = useState(false);
  const load = useCallback(
    () => salesApi.fetchDebtAccounts(outstanding),
    [outstanding],
  );
  const { items, reload } = useAdminList(load);
  const [openId, setOpenId] = useState(null);
  const [adding, setAdding] = useState(false);

  const totalOwed = (items ?? []).reduce((s, a) => s + a.balance, 0);
  const withDebt = (items ?? []).filter((a) => a.balance > 0).length;

  return (
    <>
      <PageHeader
        icon="fa-solid fa-wallet"
        title={tr("Customer Debts", "ديون العملاء")}
        subtitle={tr(
          "Customers who can pay later, and what they owe.",
          "العملاء الذين يمكنهم الدفع لاحقاً والمبالغ المستحقة عليهم.",
        )}
      >
        <Button icon="fa-solid fa-user-plus" onClick={() => setAdding(true)}>
          {tr("Allow a customer", "السماح لعميل")}
        </Button>
      </PageHeader>

      <StatsGrid>
        <StatCard
          icon="fa-solid fa-sack-dollar"
          tone="red"
          label={tr("Total owed", "إجمالي المستحق")}
          value={formatMoney(totalOwed)}
        />
        <StatCard
          icon="fa-solid fa-users"
          tone="orange"
          label={tr("Customers owing", "عملاء عليهم دين")}
          value={withDebt}
        />
        <StatCard
          icon="fa-solid fa-id-card"
          tone="blue"
          label={tr("Credit accounts", "حسابات الائتمان")}
          value={items?.length ?? 0}
        />
      </StatsGrid>

      <div className={toolbarClass}>
        <label className={styles.check}>
          <input
            type="checkbox"
            checked={outstanding}
            onChange={(e) => setOutstanding(e.target.checked)}
          />
          {tr("Only customers who owe money", "فقط العملاء الذين عليهم دين")}
        </label>
      </div>

      {items?.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-wallet"
          text={tr(
            "No credit accounts yet. Allow a customer to pay later to start.",
            "لا توجد حسابات بعد. اسمح لعميل بالدفع لاحقاً للبدء.",
          )}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Customer", "العميل")}</th>
                  <th>{tr("Owes", "المستحق")}</th>
                  <th>{tr("Credit limit", "حد الائتمان")}</th>
                  <th>{tr("Available", "المتاح")}</th>
                  <th>{tr("Last activity", "آخر حركة")}</th>
                </tr>
              </thead>
              <tbody>
                {items?.map((a) => (
                  <tr
                    key={a.userId}
                    className={tableClasses.clickableRow}
                    onClick={() => setOpenId(a.userId)}
                  >
                    <td>
                      <strong>{a.name || a.email}</strong>
                      <div className={styles.muted}>{a.email}</div>
                    </td>
                    <td className={a.balance > 0 ? styles.owed : ""}>
                      {formatMoney(a.balance)}
                    </td>
                    <td>{formatMoney(a.creditLimit)}</td>
                    <td>{formatMoney(a.availableCredit)}</td>
                    <td>
                      {a.lastTransactionAt
                        ? formatDate(a.lastTransactionAt, false)
                        : "—"}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {openId && (
        <StatementModal
          userId={openId}
          onClose={() => setOpenId(null)}
          onChanged={reload}
        />
      )}
      {adding && (
        <AllowCustomerModal
          onClose={() => setAdding(false)}
          onDone={async (userId) => {
            setAdding(false);
            await reload();
            setOpenId(userId);
          }}
        />
      )}
    </>
  );
}

// Pick a customer and give them a credit limit.
function AllowCustomerModal({ onClose, onDone }) {
  const { tr } = useLang();
  const { run, busy } = useApiAction();
  const [users, setUsers] = useState([]);
  const [userId, setUserId] = useState("");
  const [limit, setLimit] = useState("");

  useEffect(() => {
    let current = true;
    fetchUsers().then(
      (list) => current && setUsers(list.filter((u) => u.role !== "admin")),
      () => {},
    );
    return () => {
      current = false;
    };
  }, []);

  const save = async () => {
    const ok = await run(
      () => salesApi.setCreditLimit(userId, limit),
      ["Credit limit saved", "تم حفظ حد الائتمان"],
    );
    if (ok) onDone(Number(userId));
  };

  return (
    <Modal
      title={tr("Allow pay later", "السماح بالدفع لاحقاً")}
      onClose={onClose}
      footer={
        <>
          <Button variant="ghost" onClick={onClose}>
            {tr("Cancel", "إلغاء")}
          </Button>
          <Button
            icon="fa-solid fa-check"
            disabled={busy || !userId || !(Number(limit) > 0)}
            onClick={save}
          >
            {tr("Save", "حفظ")}
          </Button>
        </>
      }
    >
      <FormField label={tr("Customer", "العميل")}>
        <SelectInput
          value={userId}
          onChange={setUserId}
          options={[
            { value: "", label: tr("Choose a customer…", "اختر عميلاً…") },
            ...users.map((u) => ({
              value: String(u.userID),
              label: `${u.userFullName} · ${u.userEmail}`,
            })),
          ]}
        />
      </FormField>
      <FormField
        label={tr("Credit limit", "حد الائتمان")}
        name="limit"
        type="number"
        min="0"
        step="0.01"
        value={limit}
        onChange={(e) => setLimit(e.target.value)}
        hint={tr(
          "The most this customer may owe at one time.",
          "أقصى مبلغ يمكن أن يكون مستحقاً على العميل.",
        )}
      />
    </Modal>
  );
}

// One customer's account: change the limit, record a payment or a correction,
// and see every transaction.
function StatementModal({ userId, onClose, onChanged }) {
  const { tr, isAr } = useLang();
  const { run, busy } = useApiAction();
  const [data, setData] = useState(null);
  const [limit, setLimit] = useState("");
  const [payment, setPayment] = useState({ amount: "", note: "" });
  const [adjust, setAdjust] = useState({ amount: "", note: "" });

  const load = useCallback(() => salesApi.fetchStatement(userId), [userId]);
  const apply = (res) => {
    setData(res);
    setLimit(String(res.account.creditLimit));
  };

  useEffect(() => {
    let current = true;
    load().then(
      (res) => current && apply(res),
      () => {},
    );
    return () => {
      current = false;
    };
  }, [load]);

  const act = async (fn, success, reset) => {
    const ok = await run(fn, success);
    if (!ok) return;
    reset?.();
    apply(await load());
    onChanged();
  };

  const a = data?.account;

  return (
    <Modal
      wide
      title={a ? a.name || a.email : tr("Loading…", "جارٍ التحميل…")}
      onClose={onClose}
    >
      {a && (
        <div className={styles.statement}>
          <div className={styles.summary}>
            <div>
              <small>{tr("Owes", "المستحق")}</small>
              <strong className={a.balance > 0 ? styles.owed : ""}>
                {formatMoney(a.balance)}
              </strong>
            </div>
            <div>
              <small>{tr("Credit limit", "حد الائتمان")}</small>
              <strong>{formatMoney(a.creditLimit)}</strong>
            </div>
            <div>
              <small>{tr("Available", "المتاح")}</small>
              <strong>{formatMoney(a.availableCredit)}</strong>
            </div>
          </div>

          <div className={styles.inline}>
            <FormField
              label={tr(
                "Credit limit (0 = no pay later)",
                "حد الائتمان (0 = بدون دفع لاحق)",
              )}
              name="limit"
              type="number"
              min="0"
              step="0.01"
              value={limit}
              onChange={(e) => setLimit(e.target.value)}
            />
            <div>
              <Button
                variant="outline"
                disabled={
                  busy || limit === "" || Number(limit) === a.creditLimit
                }
                onClick={() =>
                  act(
                    () => salesApi.setCreditLimit(userId, limit),
                    ["Credit limit saved", "تم حفظ حد الائتمان"],
                  )
                }
              >
                {tr("Save limit", "حفظ الحد")}
              </Button>
            </div>
          </div>

          <div className={styles.inline}>
            <FormField
              label={tr("Payment received", "دفعة مستلمة")}
              name="payAmount"
              type="number"
              min="0"
              step="0.01"
              value={payment.amount}
              onChange={(e) =>
                setPayment((p) => ({ ...p, amount: e.target.value }))
              }
            />
            <FormField
              label={tr("Note (optional)", "ملاحظة (اختياري)")}
              name="payNote"
              value={payment.note}
              onChange={(e) =>
                setPayment((p) => ({ ...p, note: e.target.value }))
              }
            />
            <div>
              <Button
                icon="fa-solid fa-hand-holding-dollar"
                disabled={busy || !(Number(payment.amount) > 0)}
                onClick={() =>
                  act(
                    () =>
                      salesApi.recordPayment(
                        userId,
                        payment.amount,
                        payment.note,
                      ),
                    ["Payment recorded", "تم تسجيل الدفعة"],
                    () => setPayment({ amount: "", note: "" }),
                  )
                }
              >
                {tr("Record payment", "تسجيل دفعة")}
              </Button>
            </div>
          </div>

          <div className={styles.inline}>
            <FormField
              label={tr(
                "Correction (+ adds, − removes)",
                "تصحيح (+ يزيد، − ينقص)",
              )}
              name="adjAmount"
              type="number"
              step="0.01"
              value={adjust.amount}
              onChange={(e) =>
                setAdjust((p) => ({ ...p, amount: e.target.value }))
              }
            />
            <FormField
              label={tr("Reason (required)", "السبب (إلزامي)")}
              name="adjNote"
              value={adjust.note}
              onChange={(e) =>
                setAdjust((p) => ({ ...p, note: e.target.value }))
              }
            />
            <div>
              <Button
                variant="outline"
                disabled={busy || !Number(adjust.amount) || !adjust.note.trim()}
                onClick={() =>
                  act(
                    () =>
                      salesApi.recordAdjustment(
                        userId,
                        adjust.amount,
                        adjust.note,
                      ),
                    ["Correction saved", "تم حفظ التصحيح"],
                    () => setAdjust({ amount: "", note: "" }),
                  )
                }
              >
                {tr("Save correction", "حفظ التصحيح")}
              </Button>
            </div>
          </div>

          {data.transactions.length === 0 ? (
            <p className={styles.muted}>
              {tr("No transactions yet.", "لا توجد حركات بعد.")}
            </p>
          ) : (
            <div className={tableClasses.wrap}>
              <table className={tableClasses.table}>
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
                        <td className={t.amount > 0 ? styles.owed : ""}>
                          {formatMoney(t.amount)}
                        </td>
                        <td>{formatMoney(t.balanceAfter)}</td>
                        <td>{t.note}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
          {data.transactions.length >= 50 && (
            <p className={styles.muted}>
              {tr("Showing the latest 50 transactions.", "تُعرض آخر 50 حركة.")}
            </p>
          )}
        </div>
      )}
    </Modal>
  );
}
