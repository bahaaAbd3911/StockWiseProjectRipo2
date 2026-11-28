import { useCallback, useEffect, useState } from "react";

import * as emailsApi from "../../../api/emails";
import {
  Button,
  Card,
  EmptyState,
  Modal,
  PageHeader,
  Pagination,
  SearchInput,
  SelectInput,
  StatCard,
  StatsGrid,
  StatusBadge,
  tableClasses,
  toolbarClass,
} from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useLang } from "../../../hooks/useLang";
import { formatDate } from "../../../utils/format";
import styles from "./Emails.module.css";

const PAGE_SIZE = 20;
// A pending email older than this means delivery is stuck (the worker checks every few seconds).
const STUCK_MINUTES = 15;

const KINDS = {
  order_placed: ["Order received", "استلام طلب"],
  order_status: ["Order status", "حالة الطلب"],
  debt_payment: ["Payment received", "دفعة مستلمة"],
  debt_adjustment: ["Balance corrected", "تصحيح رصيد"],
  debt_reminder: ["Balance reminder", "تذكير بالرصيد"],
  back_in_stock: ["Back in stock", "عودة للمخزون"],
};

const STATUS = {
  pending: ["orange", "Pending", "بالانتظار"],
  failed: ["red", "Failed", "فشل"],
  sent: ["green", "Sent", "أُرسل"],
};

function EmailStatus({ status }) {
  const { tr } = useLang();
  const [tone, en, ar] = STATUS[status] ?? ["gray", status, status];
  return <StatusBadge tone={tone}>{tr(en, ar)}</StatusBadge>;
}

// The store's outgoing emails (order, payment, reminder, back-in-stock). They
// are queued and sent by a background worker that retries failures, so an
// email problem never breaks a checkout. This page shows whether they go out,
// and lets the admin retry or cancel them.
export default function Emails() {
  const { tr, isAr } = useLang();
  const { run, busy } = useApiAction();
  const [summary, setSummary] = useState(null);
  const [filters, setFilters] = useState({ status: "", kind: "", to: "" });
  const [search, setSearch] = useState("");
  const [page, setPage] = useState(1);
  const [list, setList] = useState(null);
  const [openId, setOpenId] = useState(null);

  // Typing in the search waits a moment before asking the server.
  useEffect(() => {
    const t = setTimeout(() => {
      setFilters((f) => (f.to === search ? f : { ...f, to: search }));
      setPage(1);
    }, 350);
    return () => clearTimeout(t);
  }, [search]);

  const load = useCallback(
    () =>
      Promise.all([
        emailsApi.fetchEmailSummary(),
        emailsApi.fetchEmails({
          ...filters,
          limit: PAGE_SIZE,
          offset: (page - 1) * PAGE_SIZE,
        }),
      ]),
    [filters, page],
  );
  // stuck: the oldest waiting email is older than STUCK_MINUTES (worked out
  // when the numbers arrive, not while rendering).
  const apply = ([s, l]) => {
    const age = s.oldestPendingAt ? (Date.now() - new Date(s.oldestPendingAt)) / 60000 : 0;
    setSummary({ ...s, stuck: age > STUCK_MINUTES });
    setList(l);
  };
  const reload = () => load().then(apply, () => {});

  useEffect(() => {
    let current = true;
    load().then(
      (res) => current && apply(res),
      () => current && setList({ items: [], total: 0 }),
    );
    return () => {
      current = false;
    };
  }, [load]);

  const setFilter = (key, value) => {
    setFilters((f) => ({ ...f, [key]: value }));
    setPage(1);
  };

  const kindLabel = (k) => {
    const [en, ar] = KINDS[k] ?? [k, k];
    return tr(en, ar);
  };
  const totalPages = Math.max(1, Math.ceil((list?.total ?? 0) / PAGE_SIZE));

  return (
    <>
      <PageHeader
        icon="fa-solid fa-envelope"
        title={tr("Emails", "رسائل البريد")}
        subtitle={tr(
          "Order, payment and stock emails sent to your customers.",
          "رسائل الطلبات والدفعات والمخزون المرسلة إلى عملائك.",
        )}
      >
        {summary?.failed > 0 && (
          <Button
            icon="fa-solid fa-rotate"
            disabled={busy}
            onClick={() =>
              run(async () => {
                await emailsApi.retryFailedEmails();
                await reload();
              }, [
                "Failed emails will be sent again",
                "ستُرسل الرسائل الفاشلة مرة أخرى",
              ])
            }
          >
            {tr(
              `Retry ${summary.failed} failed`,
              `إعادة إرسال ${summary.failed} فاشلة`,
            )}
          </Button>
        )}
      </PageHeader>

      {summary && (
        <StatsGrid>
          <StatCard
            icon="fa-solid fa-paper-plane"
            tone="green"
            label={tr("Sent in 24 hours", "أُرسلت خلال 24 ساعة")}
            value={summary.sentLast24h}
          />
          <StatCard
            icon="fa-solid fa-hourglass-half"
            tone="orange"
            label={tr("Waiting to send", "بانتظار الإرسال")}
            value={summary.pending}
          />
          <StatCard
            icon="fa-solid fa-triangle-exclamation"
            tone="red"
            label={tr("Failed", "فشلت")}
            value={summary.failed}
          />
        </StatsGrid>
      )}

      {(summary?.failed > 0 || summary?.stuck) && (
        <div className={styles.alert} role="alert">
          <i className="fa-solid fa-circle-exclamation" />
          <div>
            <strong>{tr("Emails are not going out", "الرسائل لا تُرسل")}</strong>
            <p>
              {tr(
                "Usually the email (SMTP) settings of the server are wrong or the provider is down. Ask whoever runs the server to check them, then press Retry.",
                "غالباً إعدادات البريد (SMTP) في الخادم غير صحيحة أو المزود متوقف. اطلب من مسؤول الخادم فحصها ثم اضغط إعادة الإرسال.",
              )}
            </p>
          </div>
        </div>
      )}

      <div className={toolbarClass}>
        <SearchInput
          value={search}
          onChange={setSearch}
          placeholder={tr("Search by email address", "ابحث بعنوان البريد")}
        />
        <SelectInput
          value={filters.status}
          onChange={(v) => setFilter("status", v)}
          aria-label={tr("Status", "الحالة")}
          options={[
            { value: "", label: tr("All statuses", "كل الحالات") },
            { value: "pending", label: tr("Pending", "بالانتظار") },
            { value: "failed", label: tr("Failed", "فشل") },
            { value: "sent", label: tr("Sent", "أُرسل") },
          ]}
        />
        <SelectInput
          value={filters.kind}
          onChange={(v) => setFilter("kind", v)}
          aria-label={tr("Type", "النوع")}
          options={[
            { value: "", label: tr("All types", "كل الأنواع") },
            ...Object.keys(KINDS).map((k) => ({
              value: k,
              label: kindLabel(k),
            })),
          ]}
        />
      </div>

      {list?.items.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-envelope-open"
          text={tr("No emails here.", "لا توجد رسائل هنا.")}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Date", "التاريخ")}</th>
                  <th>{tr("To", "إلى")}</th>
                  <th>{tr("Subject", "الموضوع")}</th>
                  <th>{tr("Type", "النوع")}</th>
                  <th>{tr("Status", "الحالة")}</th>
                  <th>{tr("Tries", "المحاولات")}</th>
                </tr>
              </thead>
              <tbody>
                {list?.items.map((e) => (
                  <tr
                    key={e.id}
                    className={tableClasses.clickableRow}
                    onClick={() => setOpenId(e.id)}
                  >
                    <td>{formatDate(e.createdAt, isAr, true)}</td>
                    <td>{e.to}</td>
                    <td className={styles.subject}>{e.subject}</td>
                    <td>{kindLabel(e.kind)}</td>
                    <td>
                      <EmailStatus status={e.status} />
                    </td>
                    <td>{e.attempts}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {totalPages > 1 && (
            <Pagination
              page={page}
              totalPages={totalPages}
              onChange={setPage}
              summary={tr(`${list.total} emails`, `${list.total} رسالة`)}
            />
          )}
        </Card>
      )}

      {openId && (
        <EmailModal
          id={openId}
          onClose={() => setOpenId(null)}
          onChanged={async () => {
            setOpenId(null);
            await reload();
          }}
        />
      )}
    </>
  );
}

// One email: its full text, the last error, and retry / cancel while it is not sent.
function EmailModal({ id, onClose, onChanged }) {
  const { tr, isAr } = useLang();
  const { run, busy } = useApiAction();
  const [email, setEmail] = useState(null);

  useEffect(() => {
    let current = true;
    emailsApi.fetchEmail(id).then(
      (e) => current && setEmail(e),
      () => {},
    );
    return () => {
      current = false;
    };
  }, [id]);

  const act = (fn, success) =>
    run(async () => {
      await fn(id);
      await onChanged();
    }, success);

  return (
    <Modal
      wide
      title={email?.subject ?? tr("Loading…", "جارٍ التحميل…")}
      onClose={onClose}
      footer={
        email &&
        email.status !== "sent" && (
          <>
            <Button
              variant="danger"
              icon="fa-solid fa-ban"
              disabled={busy}
              onClick={() =>
                act(emailsApi.cancelEmail, ["Email cancelled", "أُلغيت الرسالة"])
              }
            >
              {tr("Cancel email", "إلغاء الرسالة")}
            </Button>
            <Button
              icon="fa-solid fa-paper-plane"
              disabled={busy}
              onClick={() =>
                act(emailsApi.retryEmail, [
                  "The email will be sent now",
                  "ستُرسل الرسالة الآن",
                ])
              }
            >
              {tr("Send now", "إرسال الآن")}
            </Button>
          </>
        )
      }
    >
      {email && (
        <div className={styles.detail}>
          <div className={styles.meta}>
            <span>
              {tr("To", "إلى")}: <strong>{email.to}</strong>
            </span>
            <EmailStatus status={email.status} />
            <span>
              {tr("Created", "أُنشئت")}:{" "}
              {formatDate(email.createdAt, isAr, true)}
            </span>
            {email.sentAt && (
              <span>
                {tr("Sent", "أُرسلت")}: {formatDate(email.sentAt, isAr, true)}
              </span>
            )}
            {email.status === "pending" && email.nextAttemptAt && (
              <span>
                {tr("Next try", "المحاولة التالية")}:{" "}
                {formatDate(email.nextAttemptAt, isAr, true)}
              </span>
            )}
          </div>
          {email.lastError && (
            <p className={styles.error}>
              <strong>{tr("Last error", "آخر خطأ")}:</strong> {email.lastError}
            </p>
          )}
          <pre className={styles.body}>{email.body}</pre>
        </div>
      )}
    </Modal>
  );
}
