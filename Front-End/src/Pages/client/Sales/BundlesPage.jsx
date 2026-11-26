import { useEffect, useState } from "react";

import { useLang } from "../../../hooks/useLang";
import { useShop } from "../../../hooks/useShop";
import { fetchLiveBundles } from "../../../api/sales";
import { formatMoney } from "../../../utils/format";
import Header from "../../../components/client/Header/Header";
import UserSettings from "../../../components/common/UserSettings/UserSettings";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import Footer from "../../../components/common/Footer/Footer";
import cartImg from "../../../assets/images/ShoppingCart/cartLogo.webp";
import styles from "./Sales.module.css";

// /client/bundles: product sets sold together for one lower price.
export default function BundlesPage() {
  const { tr, isAr } = useLang();
  const { addBundle } = useShop();
  const [bundles, setBundles] = useState(null);
  const [adding, setAdding] = useState(null);

  useEffect(() => {
    let current = true;
    fetchLiveBundles()
      .then((list) => current && setBundles(list))
      .catch(() => current && setBundles([]));
    return () => {
      current = false;
    };
  }, []);

  const add = async (bundle) => {
    setAdding(bundle.id);
    await addBundle(bundle);
    setAdding(null);
  };

  return (
    <>
      <UserSettings />
      <SearchBar value={1} />
      <div className={styles.page}>
        <Header
          icon={"fa-solid fa-boxes-stacked"}
          title={"Bundles"}
          arbTitle={"الباقات"}
          subTitle={"Products that go together, for one lower price."}
          arbSubTitle={"منتجات تُشترى معاً بسعر أقل."}
          img={cartImg}
        />

        {bundles === null && (
          <p className={styles.muted}>{tr("Loading…", "جارٍ التحميل…")}</p>
        )}
        {bundles?.length === 0 && (
          <p className={styles.note}>
            {tr("No bundles right now.", "لا توجد باقات حالياً.")}
          </p>
        )}

        <div className={styles.grid}>
          {bundles?.map((b) => (
            <article className={styles.card} key={b.id}>
              {(b.image || b.items[0]?.image) && (
                <img src={b.image || b.items[0].image} alt="" />
              )}
              <h3>{isAr ? b.nameAr || b.name : b.name}</h3>
              {(isAr ? b.descriptionAr || b.description : b.description) && (
                <p>{isAr ? b.descriptionAr || b.description : b.description}</p>
              )}
              <ul className={styles.contents}>
                {b.items.map((i) => (
                  <li key={`${i.productId}-${i.color}`}>
                    {i.quantity} × {isAr ? i.nameAr || i.name : i.name}
                    {i.color && ` (${i.color})`}
                  </li>
                ))}
              </ul>
              <div className={styles.price}>
                <strong>{formatMoney(b.price)}</strong>
                {b.originalTotal > b.price && (
                  <s>{formatMoney(b.originalTotal)}</s>
                )}
                {b.savings > 0 && (
                  <span className={styles.good}>
                    {tr(
                      `Save ${formatMoney(b.savings)}`,
                      `وفّر ${formatMoney(b.savings)}`,
                    )}
                  </span>
                )}
              </div>
              <button
                type="button"
                className={styles.btn}
                disabled={!b.isAvailable || adding === b.id}
                onClick={() => add(b)}
              >
                <i
                  className={`fa-solid ${adding === b.id ? "fa-spinner fa-spin" : "fa-cart-shopping"}`}
                ></i>
                {b.isAvailable
                  ? tr("Add To Cart", "أضف الى سلة التسوق")
                  : b.unavailableReason || tr("Not available", "غير متاح")}
              </button>
            </article>
          ))}
        </div>
      </div>
      <Footer />
    </>
  );
}
