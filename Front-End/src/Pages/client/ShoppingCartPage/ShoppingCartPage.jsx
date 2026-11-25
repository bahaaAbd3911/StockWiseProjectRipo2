import { useContext } from "react";

import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";
import Header from "../../../components/client/Header/Header";
import EmptyMessage from "../../../components/client/EmptyMessage/EmptyMessage";
import SignInRequired from "../../../components/client/SignInRequired/SignInRequired";
import CartItemRow from "../../../components/client/Cart/CartItemRow";
import CartSummary from "../../../components/client/Cart/CartSummary";
import styles from "../../../components/client/Cart/Cart.module.css";
import cartImg from "../../../assets/images/ShoppingCart/cartLogo.png";

// Shown inside the profile page ("My Cart" tab) and at /client/shoppingList.
function ShoppingCartPage() {
  const { isSign } = useContext(AuthContext);
  const { cart } = useShop();

  if (!isSign) return <SignInRequired />;

  return (
    <div>
      <Header
        icon={"fa-solid fa-cart-shopping"}
        title={"My Shopping Cart"}
        arbTitle={"سلة التسوق"}
        subTitle={"Review your items, then continue to checkout."}
        arbSubTitle={"راجع منتجاتك ثم تابع لإتمام الشراء."}
        img={cartImg}
      />

      {cart.length === 0 ? (
        <EmptyMessage
          img={cartImg}
          title={"Your cart is empty"}
          arbTitle={"سلة التسوق فارغة"}
          subTitle={"Looks like you haven't added any products yet. Explore our collection and find something you'll love!"}
          arbSubTitle={"يبدو أنك لم تُضِف أي منتجات بعد. استكشف مجموعتنا واعثر على ما يعجبك!"}
        />
      ) : (
        <div className={styles.layout}>
          <div className={styles.list}>
            {cart.map((item) => (
              <CartItemRow key={item.id} item={item} />
            ))}
          </div>
          <CartSummary />
        </div>
      )}
    </div>
  );
}

export default ShoppingCartPage;
