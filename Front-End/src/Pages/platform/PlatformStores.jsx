import { useCallback, useEffect, useState } from "react";

import * as platformApi from "../../api/platform";
import {
  Button,
  ConfirmDialog,
  EmptyState,
  PageHeader,
  SearchInput,
  SelectInput,
  StatsGrid,
  StatCard,
  StatusBadge,
  tableClasses,
  toolbarClass,
} from "../../components/admin/ui";
import { useLang } from "../../hooks/useLang";
import { useToast } from "../../hooks/useToast";
import StoreFormModal from "./StoreFormModal";
import styles from "./Platform.module.css";

const PAGE = 20;

export default function PlatformStores({ admin, onSignOut, onUnauthorized }) {
  const { tr } = useLang();
  const toast = useToast();
  const [filters, setFilters] = useState({ search: "", status: "" });
  const [offset, setOffset] = useState(0);
  const [page, setPage] = useState({ items: [], total: 0 });
  const [editing, setEditing] = useState(null); // null | {} (new) | store
  const [confirm, setConfirm] = useState(null); // store to suspend / activate
  const [busy, setBusy] = useState(false);

  const fetchPage = useCallback(
    () => platformApi.listStores({ ...filters, offset, limit: PAGE }),
    [filters, offset],
  );

  // Refresh after an action (create, edit, suspend...).
  const load = async () => setPage(await fetchPage());

  // Load whenever the filters or the page change; ignore answers that arrive
  // after the filters changed again.
  useEffect(() => {
    let current = true;
    fetchPage().then(
      (p) => current && setPage(p),
      (err) => current && err.status === 401 && onUnauthorized(),
    );
    return () => {
      current = false;
    };
  }, [fetchPage, onUnauthorized]);

  const run = async (action, after) => {
    setBusy(true);
    try {
      const res = await action();
      if (res?.message) toast.success(res.message, res.message);
      after?.(res);
      await load();
      return true;
    } catch (err) {
      if (err.status === 401) onUnauthorized();
      toast.error(err.message, err.message);
      return false;
    } finally {
      setBusy(false);
    }
  };

  const save = (body) =>
    run(() => (editing?.id ? platformApi.updateStore(editing.id, body) : platformApi.createStore(body)), () => setEditing(null));

  const toggle = () =>
    run(
      () => (confirm.status === "active" ? platformApi.suspendStore(confirm.id) : platformApi.activateStore(confirm.id)),
      () => setConfirm(null),
    );

  const setFilter = (key) => (value) => {
    setOffset(0);
    setFilters((f) => ({ ...f, [key]: value }));
  };

  const active = page.items.filter((s) => s.status === "active").length;

  return (
    <>
      <PageHeader
        icon="fa-solid fa-store"
        title={tr("Stores", "المتاجر")}
        subtitle={tr(`Signed in as ${admin.email}`, `مسجّل الدخول باسم ${admin.email}`)}
      >
        <Button icon="fa-solid fa-plus" onClick={() => setEditing({})}>
          {tr("New store", "متجر جديد")}
        </Button>
        <Button variant="ghost" icon="fa-solid fa-right-from-bracket" onClick={onSignOut}>
          {tr("Sign out", "تسجيل الخروج")}
        </Button>
      </PageHeader>

      <StatsGrid>
        <StatCard icon="fa-solid fa-store" label={tr("Stores", "المتاجر")} value={page.total} />
        <StatCard icon="fa-solid fa-circle-check" label={tr("Active on this page", "النشطة في هذه الصفحة")} value={active} />
        <StatCard icon="fa-solid fa-circle-pause" tone="red" label={tr("Suspended on this page", "الموقوفة في هذه الصفحة")} value={page.items.length - active} />
      </StatsGrid>

      <div className={toolbarClass}>
        <SearchInput value={filters.search} onChange={setFilter("search")} placeholder={tr("Search by name or address", "ابحث بالاسم أو العنوان")} />
        <SelectInput
          value={filters.status}
          onChange={setFilter("status")}
          options={[
            { value: "", label: tr("All statuses", "كل الحالات") },
            { value: "active", label: tr("Active", "نشط") },
            { value: "suspended", label: tr("Suspended", "موقوف") },
          ]}
        />
      </div>

      {page.items.length === 0 ? (
        <EmptyState icon="fa-solid fa-store-slash" text={tr("No stores yet", "لا توجد متاجر بعد")} />
      ) : (
        <div className={tableClasses.wrap}>
          <table className={tableClasses.table}>
            <thead>
              <tr>
                <th>{tr("Store", "المتجر")}</th>
                <th>{tr("Status", "الحالة")}</th>
                <th>{tr("Admin", "المدير")}</th>
                <th>{tr("Products", "المنتجات")}</th>
                <th>{tr("Customers", "العملاء")}</th>
                <th>{tr("Orders", "الطلبات")}</th>
                <th>{tr("Revenue", "الإيرادات")}</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {page.items.map((s) => (
                <tr key={s.id}>
                  <td>
                    <strong>{s.name}</strong>
                    <br />
                    <a href={s.url} target="_blank" rel="noreferrer" className={styles.muted}>
                      {s.url.replace(/^https?:\/\//, "")}
                    </a>
                  </td>
                  <td>
                    <StatusBadge tone={s.status === "active" ? "green" : "red"}>
                      {s.status === "active" ? tr("Active", "نشط") : tr("Suspended", "موقوف")}
                    </StatusBadge>
                  </td>
                  <td>{s.admin_email ?? "—"}</td>
                  <td>{s.products_count}</td>
                  <td>{s.customers_count}</td>
                  <td>{s.orders_count}</td>
                  <td>
                    {Number(s.revenue).toFixed(2)} {s.currency}
                  </td>
                  <td className={tableClasses.rowActions}>
                    <Button variant="ghost" icon="fa-solid fa-pen" onClick={() => setEditing(s)}>
                      {tr("Edit", "تعديل")}
                    </Button>
                    <Button variant="ghost" icon="fa-solid fa-envelope" disabled={busy}
                      onClick={() => run(() => platformApi.resendInvite(s.id))}>
                      {tr("Invite again", "إعادة الدعوة")}
                    </Button>
                    <Button variant={s.status === "active" ? "danger" : "outline"}
                      icon={s.status === "active" ? "fa-solid fa-pause" : "fa-solid fa-play"} onClick={() => setConfirm(s)}>
                      {s.status === "active" ? tr("Suspend", "إيقاف") : tr("Activate", "تفعيل")}
                    </Button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {page.total > PAGE && (
        <div className={styles.pager}>
          <Button variant="ghost" disabled={offset === 0} onClick={() => setOffset((o) => Math.max(0, o - PAGE))}>
            {tr("Previous", "السابق")}
          </Button>
          <span className={styles.muted}>
            {offset + 1}–{Math.min(offset + PAGE, page.total)} / {page.total}
          </span>
          <Button variant="ghost" disabled={offset + PAGE >= page.total} onClick={() => setOffset((o) => o + PAGE)}>
            {tr("Next", "التالي")}
          </Button>
        </div>
      )}

      {editing && <StoreFormModal store={editing} busy={busy} onSave={save} onClose={() => setEditing(null)} />}

      {confirm && (
        <ConfirmDialog
          title={confirm.status === "active" ? tr("Suspend store", "إيقاف المتجر") : tr("Activate store", "تفعيل المتجر")}
          message={
            confirm.status === "active"
              ? tr(
                  `${confirm.name} will be closed for everyone (customers and its admin). Nothing is deleted; you can activate it again.`,
                  `سيُغلق ${confirm.name} أمام الجميع (العملاء ومديره). لن يُحذف شيء ويمكنك تفعيله مجدداً.`,
                )
              : tr(`${confirm.name} will be open again.`, `سيُفتح ${confirm.name} من جديد.`)
          }
          confirmLabel={confirm.status === "active" ? tr("Suspend", "إيقاف") : tr("Activate", "تفعيل")}
          onConfirm={toggle}
          onCancel={() => setConfirm(null)}
        />
      )}
    </>
  );
}
