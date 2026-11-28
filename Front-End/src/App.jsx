// Import Style
import "./App.css";

// Import Router
import { Routes, Route, Navigate } from "react-router";

// Import Context
import { AuthContext } from "./context/AuthContext";
import { Suspense, lazy, useContext, useEffect } from "react";

// Import Layouts & Pages
import ClientLayout from "./layouts/ClientLayout/ClientLayout";
import LoadingPage from "./Pages/LoadingPage/LoadingPage";
import VerifyEmail from "./Pages/account/VerifyEmail";
import ResetPassword from "./Pages/account/ResetPassword";
import { DebtLink, OrderLink, ProductLink } from "./Pages/account/EmailLinks";

// The admin panel is loaded only when an admin opens it, so shoppers do not
// download it.
const AdminLayout = lazy(() => import("./layouts/AdminLayout/AdminLayout"));
const Dashboard = lazy(() => import("./Pages/admin/Dashboard/Dashboard"));
const Products = lazy(() => import("./Pages/admin/Products/Products"));
const Categories = lazy(() => import("./Pages/admin/Categories/Categories"));
const CategoryDetails = lazy(
  () => import("./Pages/admin/CategoryDetails/CategoryDetails"),
);
const Orders = lazy(() => import("./Pages/admin/Orders/Orders"));
const Users = lazy(() => import("./Pages/admin/Users/Users"));
const UserProfile = lazy(() => import("./Pages/admin/UserProfile/UserProfile"));
const StoreSettings = lazy(
  () => import("./Pages/admin/StoreSettings/StoreSettings"),
);
const Domains = lazy(() => import("./Pages/admin/Domains/Domains"));
const Coupons = lazy(() => import("./Pages/admin/Sales/Coupons"));
const Offers = lazy(() => import("./Pages/admin/Sales/Offers"));
const Bundles = lazy(() => import("./Pages/admin/Sales/Bundles"));
const Debts = lazy(() => import("./Pages/admin/Sales/Debts"));
const WaitingList = lazy(() => import("./Pages/admin/Sales/WaitingList"));
const Emails = lazy(() => import("./Pages/admin/Emails/Emails"));

// Import Components
import Message from "./components/client/Message/Message";
import SignUp from "./components/client/Form/SignUp";
import SignIn from "./components/client/Form/SignIn";

function App() {
  const { message, setMessage } = useContext(AuthContext);

  useEffect(() => {
    if (!message.isShow) return;

    const timer = setTimeout(() => {
      setMessage({
        isShow: false,
        msgTxt: "",
        arbMsgTxt: "",
        msgBGColor: "",
        time: 0,
        reminder: false,
      });
    }, message.time);

    return () => clearTimeout(timer);
  }, [message, setMessage]);

  return (
    <div className={"appContainer"}>
      {message.isShow && (
        <Message
          mesgTxt={message.msgTxt}
          msgBGColor={message.msgBGColor}
          reminder={message.reminder}
          arbMsgTxt={message.arbMsgTxt}
        />
      )}
      <Suspense fallback={null}>
        <Routes>
          <Route path="/" element={<LoadingPage />} />
          <Route path="/client/*" element={<ClientLayout />} />

          <Route path="/admin" element={<AdminLayout />}>
            <Route index element={<Dashboard />} />
            <Route path="products" element={<Products />} />
            <Route path="categories" element={<Categories />} />
            <Route
              path="categories/:categoryId"
              element={<CategoryDetails />}
            />
            <Route path="orders" element={<Orders />} />
            <Route path="users" element={<Users />} />
            <Route path="users/:userId" element={<UserProfile />} />
            <Route path="settings" element={<StoreSettings />} />
            <Route path="domains" element={<Domains />} />
            <Route path="coupons" element={<Coupons />} />
            <Route path="offers" element={<Offers />} />
            <Route path="bundles" element={<Bundles />} />
            <Route path="debts" element={<Debts />} />
            <Route path="waiting-list" element={<WaitingList />} />
            <Route path="emails" element={<Emails />} />
            <Route path="*" element={<Navigate to="/admin" replace />} />
          </Route>

          <Route path="/signUp" element={<SignUp sign={"Up"} />} />
          <Route path="/signIn" element={<SignIn sign={"In"} />} />
          {/* Opened from emails (paths match the backend's EMAIL_VERIFY_URL / PASSWORD_RESET_URL) */}
          <Route path="/verify-email" element={<VerifyEmail />} />
          <Route path="/reset-password" element={<ResetPassword />} />
          <Route path="/orders/:orderId" element={<OrderLink />} />
          <Route path="/account/debt" element={<DebtLink />} />
          <Route path="/products/:productId" element={<ProductLink />} />
          <Route path="*" element={<Navigate to="/client/home" replace />} />
        </Routes>
      </Suspense>
    </div>
  );
}

export default App;
