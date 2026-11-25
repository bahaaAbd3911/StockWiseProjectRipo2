import { FormField, formGridClass } from "../ui";

// Out-of-stock is derived from the stock count (0 = out of stock).
export default function ProductPricingFields({ form, errors, update, tr }) {
  const numberField = (key, label, props) => (
    <FormField
      label={label}
      name={key}
      type="number"
      value={form[key]}
      error={errors[key]}
      onChange={(e) => update(key, e.target.value)}
      {...props}
    />
  );

  return (
    <div className={formGridClass}>
      {numberField("price", tr("Price (JOD)", "السعر (دينار)"), { min: 0, step: "0.01" })}
      {numberField("discountValue", tr("Discount %", "الخصم %"), { min: 0, max: 100 })}
      {numberField("count", tr("Stock count (0 = out of stock)", "الكمية (0 = نفد)"), { min: 0, step: 1 })}
      {numberField("rating", tr("Rating (0–5)", "التقييم (0–5)"), { min: 0, max: 5, step: "0.1" })}
    </div>
  );
}
