import { useEffect, useState } from "react";

import { useLang } from "../../../hooks/useLang";
import { fetchLiveOffers, fetchOfferProducts } from "../../../api/sales";
import { formatDate, formatMoney } from "../../../utils/format";
import Header from "../../../components/client/Header/Header";
import UserSettings from "../../../components/common/UserSettings/UserSettings";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import Footer from "../../../components/common/Footer/Footer";
import Products from "../../../components/client/Products/Products";
import favoriteImg from "../../../assets/images/Favorites/favoriteLogo.png";
import styles from "./Sales.module.css";

const discountText = (o, tr) =>
  o.discountType === "percentage"
    ? tr(`${o.discountValue}% off`, `خصم ${o.discountValue}%`)
    : tr(
        `${formatMoney(o.discountValue)} off`,
        `خصم ${formatMoney(o.discountValue)}`,
      );

// /client/offers: the store's live offers; pick one to see its products.
export default function OffersPage() {
  const { tr, isAr } = useLang();
  const [offers, setOffers] = useState(null);
  const [selected, setSelected] = useState(null);
  const [products, setProducts] = useState(null);

  useEffect(() => {
    let current = true;
    fetchLiveOffers()
      .then((list) => {
        if (!current) return;
        setOffers(list);
        setSelected(list[0]?.id ?? null);
      })
      .catch(() => current && setOffers([]));
    return () => {
      current = false;
    };
  }, []);

  useEffect(() => {
    if (!selected) return;
    let current = true;
    fetchOfferProducts(selected)
      .then((res) => current && setProducts(res.products))
      .catch(() => current && setProducts([]));
    return () => {
      current = false;
    };
  }, [selected]);

  const offer = offers?.find((o) => o.id === selected);
  const pick = (id) => {
    if (id === selected) return;
    setProducts(null);
    setSelected(id);
  };

  return (
    <>
      <UserSettings />
      <SearchBar value={1} />
      <div className={styles.page}>
        <Header
          icon={"fa-solid fa-tags"}
          title={"Offers"}
          arbTitle={"العروض"}
          subTitle={
            "Limited-time discounts. Prices below already include the offer."
          }
          arbSubTitle={"خصومات لفترة محدودة. الأسعار المعروضة تشمل الخصم."}
          img={favoriteImg}
        />

        {offers === null && (
          <p className={styles.muted}>{tr("Loading…", "جارٍ التحميل…")}</p>
        )}
        {offers?.length === 0 && (
          <p className={styles.note}>
            {tr(
              "No offers right now. Check back soon!",
              "لا توجد عروض حالياً. عُد قريباً!",
            )}
          </p>
        )}

        {offers?.length > 0 && (
          <div className={styles.offerHead}>
            {offers.map((o) => (
              <button
                type="button"
                key={o.id}
                className={`${styles.chip} ${o.id === selected ? styles.chipOn : ""}`}
                onClick={() => pick(o.id)}
              >
                {isAr ? o.nameAr || o.name : o.name} · {discountText(o, tr)}
              </button>
            ))}
          </div>
        )}

        {offer && (
          <p className={styles.note}>
            {isAr
              ? offer.descriptionAr || offer.description
              : offer.description}
            {offer.endsAt &&
              ` ${tr("Ends", "ينتهي")} ${formatDate(offer.endsAt, isAr, true)}.`}
          </p>
        )}

        {offer && products === null && (
          <p className={styles.muted}>{tr("Loading…", "جارٍ التحميل…")}</p>
        )}
        {offer && products && <Products productsProp={products} />}
      </div>
      <Footer />
    </>
  );
}
