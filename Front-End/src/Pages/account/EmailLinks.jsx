import { useContext, useEffect } from "react";
import { Navigate, useParams } from "react-router";

import { AuthContext } from "../../context/AuthContext";

// The backend's emails link to short, stable addresses (backend/internal/notify/templates.go):
//   /orders/{id}     order received / status changed  -> profile, "My Orders" tab
//   /account/debt    payment, correction, reminder    -> profile, "My Balance" tab
//   /products/{id}   back in stock                    -> the product page
// These routes forward them to the screens that show that information.

function ProfileTab({ tab }) {
  const { openProfileTab } = useContext(AuthContext);
  useEffect(() => openProfileTab(tab), [openProfileTab, tab]);
  return <Navigate to="/client/profile" replace />;
}

export const OrderLink = () => <ProfileTab tab={1} />;
export const DebtLink = () => <ProfileTab tab={4} />;

export function ProductLink() {
  const { productId } = useParams();
  return <Navigate to={`/client/ProductDetails/${productId}`} replace />;
}
