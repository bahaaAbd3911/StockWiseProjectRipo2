import { useState } from "react";

import * as salesApi from "../../../api/sales";
import {
  Button,
  Card,
  ConfirmDialog,
  EmptyState,
  FormField,
  IconButton,
  Modal,
  PageHeader,
  SelectInput,
  formGridClass,
  tableClasses,
} from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import { formatDate } from "../../../utils/format";
import {
  ActiveToggle,
  DateRangeFields,
  SaleStatus,
  ScopeFields,
} from "./shared";
import {
  discountLabel,
  toLocalInput,
  useAdminList,
  useFormModal,
} from "./salesUtils";

const toForm = (o) => ({
  name: o?.name ?? "",
  nameAr: o?.nameAr ?? "",
  description: o?.description ?? "",
  descriptionAr: o?.descriptionAr ?? "",
  discountType: o?.discountType ?? "percentage",
  discountValue: o?.discountValue ?? "",
  scope: o?.scope ?? "products",
  category: o?.category ?? "",
  productIds: o?.productIds ?? [],
  // Offers always run for a set time; a new one defaults to the next 7 days.
  startsAt: toLocalInput(o?.startsAt ?? new Date().toISOString()),
  endsAt: toLocalInput(
    o?.endsAt ?? new Date(Date.now() + 7 * 864e5).toISOString(),
  ),
  isActive: o?.isActive ?? true,
});

// Time-limited price cuts on products or a whole category. Prices in the shop
// already show the offer; customers do not need a code.
export default function Offers() {
  const { tr, isAr } = useLang();
  const { run } = useApiAction();
  const { reload: reloadCatalog } = useCatalog();
  const { items, reload } = useAdminList(salesApi.fetchAdminOffers);
  // Offer prices show on the products, so the catalog is reloaded too.
  const reloadAll = () => Promise.all([reload(), reloadCatalog()]);
  const modal = useFormModal(toForm, reloadAll);
  const [deleting, setDeleting] = useState(null);
  const { form, set, errors } = modal;

  const period = (o) =>
    [
      o.startsAt && formatDate(o.startsAt, isAr),
      o.endsAt && formatDate(o.endsAt, isAr),
    ]
      .filter(Boolean)
      .join(" → ") || tr("Always", "دائماً");

  return (
    <>
      <PageHeader
        icon="fa-solid fa-tags"
        title={tr("Offers", "العروض")}
        subtitle={tr(
          "Price cuts shown directly on products — no code needed.",
          "تخفيضات تظهر مباشرة على المنتجات دون رمز.",
        )}
      >
        <Button icon="fa-solid fa-plus" onClick={() => modal.show(null)}>
          {tr("Add offer", "إضافة عرض")}
        </Button>
      </PageHeader>

      {items?.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-tags"
          text={tr("No offers yet.", "لا توجد عروض بعد.")}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Name", "الاسم")}</th>
                  <th>{tr("Discount", "الخصم")}</th>
                  <th>{tr("Applies to", "ينطبق على")}</th>
                  <th>{tr("Period", "المدة")}</th>
                  <th>{tr("Status", "الحالة")}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items?.map((o) => (
                  <tr
                    key={o.id}
                    className={tableClasses.clickableRow}
                    onClick={() => modal.show(o)}
                  >
                    <td>
                      <strong>{isAr ? o.nameAr || o.name : o.name}</strong>
                    </td>
                    <td>{discountLabel(o)}</td>
                    <td>
                      {o.scope === "category"
                        ? o.category
                        : tr(
                            `${o.productIds.length} products`,
                            `${o.productIds.length} منتجات`,
                          )}
                    </td>
                    <td>{period(o)}</td>
                    <td>
                      <SaleStatus status={o.status} />
                    </td>
                    <td>
                      <div className={tableClasses.rowActions}>
                        <IconButton
                          icon="fa-solid fa-pen"
                          label={tr("Edit", "تعديل")}
                          onClick={() => modal.show(o)}
                        />
                        <IconButton
                          icon="fa-solid fa-trash"
                          danger
                          label={tr("Delete", "حذف")}
                          onClick={() => setDeleting(o)}
                        />
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Card>
      )}

      {modal.open && (
        <Modal
          wide
          title={
            modal.open.id
              ? tr("Edit offer", "تعديل العرض")
              : tr("New offer", "عرض جديد")
          }
          onClose={modal.close}
          footer={
            <>
              <Button variant="ghost" onClick={modal.close}>
                {tr("Cancel", "إلغاء")}
              </Button>
              <Button
                icon="fa-solid fa-floppy-disk"
                disabled={modal.saving}
                onClick={() =>
                  modal.save(salesApi.saveOffer, [
                    "Offer saved",
                    "تم حفظ العرض",
                  ])
                }
              >
                {tr("Save", "حفظ")}
              </Button>
            </>
          }
        >
          <div className={formGridClass}>
            <FormField
              label={tr("Name (English)", "الاسم (إنجليزي)")}
              name="name"
              value={form.name}
              onChange={(e) => set("name", e.target.value)}
              error={errors?.name}
            />
            <FormField
              label={tr("Name (Arabic)", "الاسم (عربي)")}
              name="nameAr"
              dir="rtl"
              value={form.nameAr}
              onChange={(e) => set("nameAr", e.target.value)}
              error={errors?.name_ar}
            />
            <FormField
              label={tr("Discount type", "نوع الخصم")}
              error={errors?.discount_type}
            >
              <SelectInput
                value={form.discountType}
                onChange={(v) => set("discountType", v)}
                options={[
                  {
                    value: "percentage",
                    label: tr("Percentage (%)", "نسبة (%)"),
                  },
                  {
                    value: "fixed",
                    label: tr("Fixed amount per item", "مبلغ ثابت لكل قطعة"),
                  },
                ]}
              />
            </FormField>
            <FormField
              label={
                form.discountType === "percentage"
                  ? tr("Percent off", "نسبة الخصم")
                  : tr("Amount off", "مبلغ الخصم")
              }
              name="discountValue"
              type="number"
              min="0"
              step="0.01"
              value={form.discountValue}
              onChange={(e) => set("discountValue", e.target.value)}
              error={errors?.discount_value}
            />
            <ScopeFields form={form} set={set} errors={errors} />
            <DateRangeFields form={form} set={set} errors={errors} required />
            <FormField
              label={tr("Description (English)", "الوصف (إنجليزي)")}
              name="description"
              value={form.description}
              onChange={(e) => set("description", e.target.value)}
            />
            <FormField
              label={tr("Description (Arabic)", "الوصف (عربي)")}
              name="descriptionAr"
              dir="rtl"
              value={form.descriptionAr}
              onChange={(e) => set("descriptionAr", e.target.value)}
            />
            <ActiveToggle form={form} set={set} />
          </div>
        </Modal>
      )}

      {deleting && (
        <ConfirmDialog
          title={tr("Delete offer", "حذف العرض")}
          message={tr(
            `Delete "${deleting.name}"? Prices go back to normal.`,
            `حذف "${deleting.nameAr || deleting.name}"؟ ستعود الأسعار لوضعها الطبيعي.`,
          )}
          onConfirm={() =>
            run(async () => {
              await salesApi.deleteOffer(deleting.id);
              setDeleting(null);
              await reloadAll();
            }, ["Offer deleted", "تم حذف العرض"])
          }
          onCancel={() => setDeleting(null)}
        />
      )}
    </>
  );
}
