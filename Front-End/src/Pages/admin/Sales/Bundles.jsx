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
  inputClass,
  tableClasses,
} from "../../../components/admin/ui";
import { useApiAction } from "../../../hooks/useApiAction";
import { useCatalog } from "../../../hooks/useCatalog";
import { useLang } from "../../../hooks/useLang";
import { formatMoney } from "../../../utils/format";
import { ActiveToggle, DateRangeFields, SaleStatus } from "./shared";
import { toLocalInput, useAdminList, useFormModal } from "./salesUtils";
import styles from "./Sales.module.css";

const toForm = (b) => ({
  name: b?.name ?? "",
  nameAr: b?.nameAr ?? "",
  description: b?.description ?? "",
  descriptionAr: b?.descriptionAr ?? "",
  image: b?.image ?? "",
  price: b?.price ?? "",
  startsAt: toLocalInput(b?.startsAt),
  endsAt: toLocalInput(b?.endsAt),
  isActive: b?.isActive ?? true,
  items: b?.items?.map((i) => ({
    productId: i.productId,
    quantity: i.quantity,
    color: i.color,
  })) ?? [
    { productId: "", quantity: 1, color: "" },
    { productId: "", quantity: 1, color: "" },
  ],
});

// The status a bundle has right now (the API gives is_active and the dates).
const bundleStatus = (b) => {
  const now = Date.now();
  if (!b.isActive) return "disabled";
  if (b.startsAt && new Date(b.startsAt) > now) return "scheduled";
  if (b.endsAt && new Date(b.endsAt) <= now) return "expired";
  return "live";
};

// Several products sold together for one price.
export default function Bundles() {
  const { tr, isAr } = useLang();
  const { run } = useApiAction();
  const { products } = useCatalog();
  const { items, reload } = useAdminList(salesApi.fetchAdminBundles);
  const modal = useFormModal(toForm, reload);
  const [deleting, setDeleting] = useState(null);
  const { form, set, errors } = modal;

  const productOptions = [
    { value: "", label: tr("Choose a product…", "اختر منتجاً…") },
    ...products.map((p) => ({
      value: String(p.id),
      label: isAr ? p.nameAr || p.name : p.name,
    })),
  ];
  const setItem = (index, key, value) =>
    set(
      "items",
      form.items.map((it, i) =>
        i === index
          ? { ...it, [key]: value, ...(key === "productId" && { color: "" }) }
          : it,
      ),
    );
  // What the products would cost on their own (helps pick a bundle price).
  const separateTotal = form
    ? form.items.reduce((sum, it) => {
        const p = products.find((x) => x.id === Number(it.productId));
        return sum + (p ? p.finalPrice * (Number(it.quantity) || 0) : 0);
      }, 0)
    : 0;

  return (
    <>
      <PageHeader
        icon="fa-solid fa-boxes-stacked"
        title={tr("Bundles", "الباقات")}
        subtitle={tr(
          "Sell several products together for one price.",
          "بِع عدة منتجات معاً بسعر واحد.",
        )}
      >
        <Button icon="fa-solid fa-plus" onClick={() => modal.show(null)}>
          {tr("Add bundle", "إضافة باقة")}
        </Button>
      </PageHeader>

      {items?.length === 0 ? (
        <EmptyState
          icon="fa-solid fa-boxes-stacked"
          text={tr("No bundles yet.", "لا توجد باقات بعد.")}
        />
      ) : (
        <Card>
          <div className={tableClasses.wrap}>
            <table className={tableClasses.table}>
              <thead>
                <tr>
                  <th>{tr("Name", "الاسم")}</th>
                  <th>{tr("Contents", "المحتويات")}</th>
                  <th>{tr("Price", "السعر")}</th>
                  <th>{tr("Separately", "منفصلة")}</th>
                  <th>{tr("Can sell", "المتاح")}</th>
                  <th>{tr("Status", "الحالة")}</th>
                  <th></th>
                </tr>
              </thead>
              <tbody>
                {items?.map((b) => (
                  <tr
                    key={b.id}
                    className={tableClasses.clickableRow}
                    onClick={() => modal.show(b)}
                  >
                    <td>
                      <strong>{isAr ? b.nameAr || b.name : b.name}</strong>
                    </td>
                    <td>
                      {b.items
                        .map(
                          (i) =>
                            `${i.quantity} × ${isAr ? i.nameAr || i.name : i.name}`,
                        )
                        .join(", ")}
                    </td>
                    <td>{formatMoney(b.price)}</td>
                    <td>{formatMoney(b.originalTotal)}</td>
                    <td>{b.availableStock}</td>
                    <td>
                      <SaleStatus status={bundleStatus(b)} />
                    </td>
                    <td>
                      <div className={tableClasses.rowActions}>
                        <IconButton
                          icon="fa-solid fa-pen"
                          label={tr("Edit", "تعديل")}
                          onClick={() => modal.show(b)}
                        />
                        <IconButton
                          icon="fa-solid fa-trash"
                          danger
                          label={tr("Delete", "حذف")}
                          onClick={() => setDeleting(b)}
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
              ? tr("Edit bundle", "تعديل الباقة")
              : tr("New bundle", "باقة جديدة")
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
                  modal.save(salesApi.saveBundle, [
                    "Bundle saved",
                    "تم حفظ الباقة",
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

            <div className={styles.full}>
              <FormField
                label={tr("Products in the bundle", "منتجات الباقة")}
                error={errors?.items}
              >
                <div className={styles.items}>
                  {form.items.map((it, index) => {
                    const product = products.find(
                      (p) => p.id === Number(it.productId),
                    );
                    const colors = product?.availableColors ?? [];
                    return (
                      <div className={styles.itemRow} key={index}>
                        <SelectInput
                          value={String(it.productId)}
                          onChange={(v) => setItem(index, "productId", v)}
                          options={productOptions}
                          aria-label={tr("Product", "المنتج")}
                        />
                        <input
                          type="number"
                          min="1"
                          max="100"
                          className={inputClass}
                          value={it.quantity}
                          onChange={(e) =>
                            setItem(index, "quantity", e.target.value)
                          }
                          aria-label={tr("Quantity", "الكمية")}
                        />
                        {colors.length > 0 ? (
                          <SelectInput
                            value={it.color}
                            onChange={(v) => setItem(index, "color", v)}
                            options={[
                              { value: "", label: tr("Any color", "أي لون") },
                              ...colors.map((c) => ({
                                value: c.name,
                                label: isAr ? c.nameAr : c.name,
                              })),
                            ]}
                            aria-label={tr("Color", "اللون")}
                          />
                        ) : (
                          <span />
                        )}
                        <IconButton
                          icon="fa-solid fa-xmark"
                          danger
                          label={tr("Remove", "إزالة")}
                          onClick={() =>
                            set(
                              "items",
                              form.items.filter((_, i) => i !== index),
                            )
                          }
                        />
                      </div>
                    );
                  })}
                  <div>
                    <Button
                      variant="outline"
                      icon="fa-solid fa-plus"
                      disabled={form.items.length >= 20}
                      onClick={() =>
                        set("items", [
                          ...form.items,
                          { productId: "", quantity: 1, color: "" },
                        ])
                      }
                    >
                      {tr("Add product", "إضافة منتج")}
                    </Button>
                  </div>
                </div>
              </FormField>
            </div>

            <FormField
              label={tr("Bundle price", "سعر الباقة")}
              name="price"
              type="number"
              min="0"
              step="0.01"
              value={form.price}
              onChange={(e) => set("price", e.target.value)}
              error={errors?.price}
              hint={
                separateTotal > 0
                  ? tr(
                      `Bought separately: ${formatMoney(separateTotal)}`,
                      `منفصلة: ${formatMoney(separateTotal)}`,
                    )
                  : undefined
              }
            />
            <FormField
              label={tr("Image URL (optional)", "رابط الصورة (اختياري)")}
              name="image"
              value={form.image}
              placeholder="https://…"
              onChange={(e) => set("image", e.target.value)}
              error={errors?.image_url}
              hint={tr(
                "Empty: the first product's photo is used.",
                "فارغ: تُستخدم صورة المنتج الأول.",
              )}
            />
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
          title={tr("Delete bundle", "حذف الباقة")}
          message={tr(
            `Delete "${deleting.name}"? It is also removed from carts.`,
            `حذف "${deleting.nameAr || deleting.name}"؟ ستُزال من السلال أيضاً.`,
          )}
          onConfirm={() =>
            run(async () => {
              await salesApi.deleteBundle(deleting.id);
              setDeleting(null);
              await reload();
            }, ["Bundle deleted", "تم حذف الباقة"])
          }
          onCancel={() => setDeleting(null)}
        />
      )}
    </>
  );
}
