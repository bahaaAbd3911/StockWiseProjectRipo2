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
import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";
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

const toForm = (c) => ({
  code: c?.code ?? "",
  description: c?.description ?? "",
  descriptionAr: c?.descriptionAr ?? "",
  discountType: c?.discountType ?? "percentage",
  discountValue: c?.discountValue ?? "",
  scope: c?.scope ?? "all",
  category: c?.category ?? "",
  productIds: c?.productIds ?? [],
  minOrder: c?.minOrder || "",
  maxDiscount: c?.maxDiscount ?? "",
  usageLimit: c?.usageLimit ?? "",
  perCustomerLimit: c?.perCustomerLimit ?? "",
  startsAt: toLocalInput(c?.startsAt),
  endsAt: toLocalInput(c?.endsAt),
  isActive: c?.isActive ?? true,
});

// Codes customers type in the cart for a discount.
export default function Coupons() {
  const { tr } = useLang();
  const { run } = useApiAction();
  const { items, reload } = useAdminList(salesApi.fetchAdminCoupons);
  const modal = useFormModal(toForm, reload);
  const [deleting, setDeleting] = useState(null);
  const { form, set, errors } = modal;

  return (
    <>
      <PageHeader
        icon="fa-solid fa-ticket"
        title={tr("Coupons", "القسائم")}
        subtitle={tr(
          "Discount codes customers enter in their cart.",
          "رموز خصم يُدخلها العملاء في السلة.",
        )}
      >
        <Button icon="fa-solid fa-plus" onClick={() => modal.show(null)}>
          {tr("Add coupon", "إضافة قسيمة")}
        </Button>
      </PageHeader>

      {items?.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-ticket"
          text={tr("No coupons yet.", "لا توجد قسائم بعد.")}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Code", "الرمز")}</th>
                  <th>{tr("Discount", "الخصم")}</th>
                  <th>{tr("Applies to", "ينطبق على")}</th>
                  <th>{tr("Used", "الاستخدام")}</th>
                  <th>{tr("Given away", "إجمالي الخصم")}</th>
                  <th>{tr("Status", "الحالة")}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items?.map((c) => (
                  <tr
                    key={c.id}
                    className={tableClasses.clickableRow}
                    onClick={() => modal.show(c)}
                  >
                    <td>
                      <strong>{c.code}</strong>
                    </td>
                    <td>
                      {discountLabel(c)}
                      {c.minOrder > 0 &&
                        ` · ${tr("min", "حد أدنى")} ${formatMoney(c.minOrder)}`}
                    </td>
                    <td>
                      {c.scope === "all"
                        ? tr("Whole order", "الطلب كاملاً")
                        : c.scope === "category"
                          ? c.category
                          : tr(
                              `${c.productIds.length} products`,
                              `${c.productIds.length} منتجات`,
                            )}
                    </td>
                    <td>
                      {c.timesUsed}
                      {c.usageLimit !== "" && ` / ${c.usageLimit}`}
                    </td>
                    <td>{formatMoney(c.totalDiscount)}</td>
                    <td>
                      <SaleStatus status={c.status} />
                    </td>
                    <td>
                      <div className={tableClasses.rowActions}>
                        <IconButton
                          icon="fa-solid fa-pen"
                          label={tr("Edit", "تعديل")}
                          onClick={() => modal.show(c)}
                        />
                        <IconButton
                          icon="fa-solid fa-trash"
                          danger
                          label={tr("Delete", "حذف")}
                          onClick={() => setDeleting(c)}
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
              ? tr("Edit coupon", "تعديل القسيمة")
              : tr("New coupon", "قسيمة جديدة")
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
                  modal.save(salesApi.saveCoupon, [
                    "Coupon saved",
                    "تم حفظ القسيمة",
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
              label={tr("Code", "الرمز")}
              name="code"
              value={form.code}
              placeholder="SUMMER25"
              onChange={(e) => set("code", e.target.value.toUpperCase())}
              error={errors?.code}
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
                  { value: "fixed", label: tr("Fixed amount", "مبلغ ثابت") },
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
            {form.discountType === "percentage" && (
              <FormField
                label={tr("Maximum discount (optional)", "أقصى خصم (اختياري)")}
                name="maxDiscount"
                type="number"
                min="0"
                step="0.01"
                value={form.maxDiscount}
                onChange={(e) => set("maxDiscount", e.target.value)}
                error={errors?.max_discount}
              />
            )}
            <FormField
              label={tr(
                "Minimum order (optional)",
                "الحد الأدنى للطلب (اختياري)",
              )}
              name="minOrder"
              type="number"
              min="0"
              step="0.01"
              value={form.minOrder}
              onChange={(e) => set("minOrder", e.target.value)}
              error={errors?.min_order_amount}
            />
            <FormField
              label={tr(
                "Total uses (empty = unlimited)",
                "عدد الاستخدامات (فارغ = غير محدود)",
              )}
              name="usageLimit"
              type="number"
              min="1"
              value={form.usageLimit}
              onChange={(e) => set("usageLimit", e.target.value)}
              error={errors?.usage_limit}
            />
            <FormField
              label={tr(
                "Uses per customer (empty = unlimited)",
                "لكل عميل (فارغ = غير محدود)",
              )}
              name="perCustomerLimit"
              type="number"
              min="1"
              value={form.perCustomerLimit}
              onChange={(e) => set("perCustomerLimit", e.target.value)}
              error={errors?.per_customer_limit}
            />
            <ScopeFields form={form} set={set} errors={errors} allowAll />
            <DateRangeFields form={form} set={set} errors={errors} />
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
          title={tr("Delete coupon", "حذف القسيمة")}
          message={tr(
            `Delete ${deleting.code}? Orders that already used it keep their discount.`,
            `حذف ${deleting.code}؟ الطلبات التي استخدمته تحتفظ بخصمها.`,
          )}
          onConfirm={() =>
            run(async () => {
              await salesApi.deleteCoupon(deleting.id);
              setDeleting(null);
              await reload();
            }, ["Coupon deleted", "تم حذف القسيمة"])
          }
          onCancel={() => setDeleting(null)}
        />
      )}
    </>
  );
}
