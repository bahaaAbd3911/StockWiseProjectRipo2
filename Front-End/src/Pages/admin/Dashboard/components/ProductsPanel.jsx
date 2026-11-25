import { FaEye } from "react-icons/fa";
import { useNavigate } from "react-router";

import { finalPrice, formatMoney } from "../../../../utils/format";
import styles from "../Dashboard.module.css";

// rows: [{ product, sold }] — best sellers computed from real orders.
export default function ProductsPanel({ t, rows, isAr }) {
  const navigate = useNavigate();

  return (
    <section className={styles.panel}>
      <div className={styles.panelHeader}>
        <h3>{t.topProducts}</h3>
        <button className={styles.viewAll} onClick={() => navigate("/admin/products")}>
          {t.viewAll} <FaEye />
        </button>
      </div>

      <div className={styles.productList}>
        {rows.length === 0 && <p className={styles.panelSubtitle}>{t.noData}</p>}
        {rows.map(({ product, sold }) => (
          <div className={styles.productItem} key={product.id}>
            <span className={styles.productImage}>
              <img src={product.image} alt="" style={{ width: "100%", height: "100%", objectFit: "cover", borderRadius: "inherit" }} />
            </span>
            <div className={styles.productInfo}>
              <strong>{isAr ? product.nameAr : product.name}</strong>
              <small>{isAr ? product.categoryAr : product.category}</small>
            </div>
            <div className={styles.productNumbers}>
              <strong>
                {sold} {t.sold}
              </strong>
              <small>
                {product.count} {t.stock}
              </small>
            </div>
            <strong className={styles.productPrice}>{formatMoney(finalPrice(product))}</strong>
          </div>
        ))}
      </div>
    </section>
  );
}
