import {
  Button,
  FormField,
  IconButton,
  SelectInput,
} from "../../../../components/admin/ui";
import { jordanianCitiesList } from "../../../../data/jordanLocations";
import { useLang } from "../../../../hooks/useLang";
import SettingsCard from "./SettingsCard";
import styles from "../StoreSettings.module.css";

// What delivery costs. The backend charges it at checkout from the delivery
// city (pricing.ShippingFee): a city's own fee, else the standard fee, and
// nothing once the items reach "free over". No shipping settings = free.
// Cities come from the same list the checkout form offers, so the names match.

const num = (v) => (v === "" ? 0 : Math.max(0, Number(v) || 0));

export default function ShippingCard({ shipping, onChange, tr }) {
  const { isAr } = useLang();
  const s = shipping ?? { fee: 0, freeOver: null, cities: [] };
  const cities = s.cities ?? [];
  const set = (patch) => onChange({ ...s, cities, ...patch });
  const unused = jordanianCitiesList.filter(
    (c) => !cities.some((x) => x.city === c.cityName),
  );
  const label = (name) => {
    const c = jordanianCitiesList.find((x) => x.cityName === name);
    return isAr && c ? c.arbCityName : name;
  };

  return (
    <SettingsCard
      icon="fa-solid fa-truck"
      title={tr("Shipping", "الشحن")}
      subtitle={tr(
        "What customers pay for delivery. Charged at checkout from the delivery city.",
        "ما يدفعه العملاء للتوصيل. يُحسب عند إتمام الطلب حسب مدينة التوصيل.",
      )}
    >
      {!shipping ? (
        <div className={styles.shippingOff}>
          <p>
            {tr(
              "Shipping is free for every order.",
              "الشحن مجاني لكل الطلبات.",
            )}
          </p>
          <Button
            icon="fa-solid fa-plus"
            onClick={() => onChange({ fee: 0, freeOver: null, cities: [] })}
          >
            {tr("Charge for shipping", "إضافة رسوم شحن")}
          </Button>
        </div>
      ) : (
        <>
          <div className={styles.shippingGrid}>
            <FormField
              label={tr("Standard fee per order", "الرسوم العادية لكل طلب")}
              name="shippingFee"
              type="number"
              min="0"
              step="0.01"
              value={s.fee}
              onChange={(e) => set({ fee: num(e.target.value) })}
            />
            <FormField
              label={tr(
                "Free shipping from (optional)",
                "شحن مجاني ابتداءً من (اختياري)",
              )}
              name="freeOver"
              type="number"
              min="0"
              step="0.01"
              value={s.freeOver ?? ""}
              placeholder={tr("Never free", "غير مجاني أبداً")}
              hint={tr(
                "Items total after the coupon",
                "مجموع المنتجات بعد القسيمة",
              )}
              onChange={(e) =>
                set({
                  freeOver: e.target.value === "" ? null : num(e.target.value),
                })
              }
            />
          </div>

          <h4 className={styles.subTitle}>
            {tr("Cities with their own fee", "مدن برسوم خاصة")}
          </h4>
          {cities.length === 0 && (
            <p className={styles.muted}>
              {tr(
                "Every city pays the standard fee.",
                "كل المدن تدفع الرسوم العادية.",
              )}
            </p>
          )}
          {cities.map((c, i) => (
            <div className={styles.cityRow} key={c.city}>
              <span>{label(c.city)}</span>
              <input
                type="number"
                min="0"
                step="0.01"
                className={styles.cityFee}
                aria-label={tr("Fee", "الرسوم")}
                value={c.fee}
                onChange={(e) =>
                  set({
                    cities: cities.map((x, j) =>
                      j === i ? { ...x, fee: num(e.target.value) } : x,
                    ),
                  })
                }
              />
              <IconButton
                icon="fa-solid fa-xmark"
                danger
                label={tr("Remove", "إزالة")}
                onClick={() =>
                  set({ cities: cities.filter((_, j) => j !== i) })
                }
              />
            </div>
          ))}
          {unused.length > 0 && (
            <div className={styles.cityAdd}>
              <SelectInput
                value=""
                aria-label={tr("Add a city", "إضافة مدينة")}
                onChange={(city) =>
                  city && set({ cities: [...cities, { city, fee: s.fee }] })
                }
                options={[
                  { value: "", label: tr("+ Add a city…", "+ إضافة مدينة…") },
                  ...unused.map((c) => ({
                    value: c.cityName,
                    label: isAr ? c.arbCityName : c.cityName,
                  })),
                ]}
              />
            </div>
          )}

          <Button
            variant="ghost"
            icon="fa-solid fa-gift"
            onClick={() => onChange(null)}
          >
            {tr("Make shipping free for everyone", "جعل الشحن مجانياً للجميع")}
          </Button>
        </>
      )}
    </SettingsCard>
  );
}
