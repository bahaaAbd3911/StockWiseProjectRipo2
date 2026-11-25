import { useLang } from "../../../hooks/useLang";
import { Card, EmptyState, tableClasses } from "../ui";
import ProductRow from "./ProductRow";

export default function ProductsTable({ products, onEdit, onDelete }) {
  const { tr, isAr } = useLang();

  if (products.length === 0) {
    return (
      <Card>
        <EmptyState text={tr("No products found", "لا توجد منتجات")} />
      </Card>
    );
  }

  const headers = [
    tr("Product", "المنتج"),
    tr("Category", "التصنيف"),
    tr("Price", "السعر"),
    tr("Stock", "المخزون"),
    tr("Rating", "التقييم"),
    tr("Status", "الحالة"),
    tr("Actions", "الإجراءات"),
  ];

  return (
    <Card className={tableClasses.wrap}>
      <table className={tableClasses.table}>
        <thead>
          <tr>{headers.map((h) => <th key={h}>{h}</th>)}</tr>
        </thead>
        <tbody>
          {products.map((product) => (
            <ProductRow
              key={product.id}
              product={product}
              isAr={isAr}
              tr={tr}
              onEdit={onEdit}
              onDelete={onDelete}
            />
          ))}
        </tbody>
      </table>
    </Card>
  );
}
