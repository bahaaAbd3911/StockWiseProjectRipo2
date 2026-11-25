import { Card, EmptyState } from "../../../../components/admin/ui";
import { finalPrice, formatMoney } from "../../../../utils/format";
import styles from "../UserProfile.module.css";

export default function UserFavorites({ products, isAr, tr }) {
  return (
    <>
      <h2 className={styles.sectionTitle}>{tr("Favorite Products", "المنتجات المفضلة")}</h2>

      {products.length === 0 ? (
        <Card>
          <EmptyState icon="fa-regular fa-heart" text={tr("No favorite products yet", "لا توجد منتجات مفضلة")} />
        </Card>
      ) : (
        <div className={styles.favorites}>
          {products.map((p) => (
            <Card key={p.id} className={styles.favorite}>
              <img src={p.image} alt="" />
              <div>
                <p>{isAr ? p.nameAr : p.name}</p>
                <small>{formatMoney(finalPrice(p))}</small>
              </div>
            </Card>
          ))}
        </div>
      )}
    </>
  );
}
