// Import Style
import "./App.css";

// Import Router
import { Routes, Route, Navigate } from "react-router";

// Import Context
import { AuthContext } from "./context/AuthContext";
import { useContext, useEffect } from "react";

// Import Layouts & Pages
import ClientLayout from "./layouts/ClientLayout/ClientLayout";
import AdminLayout from "./layouts/AdminLayout/AdminLayout";
import LoadingPage from "./Pages/LoadingPage/LoadingPage";
import Dashboard from "./Pages/admin/Dashboard/Dashboard";
import Products from "./Pages/admin/Products/Products";
import Categories from "./Pages/admin/Categories/Categories";
import CategoryDetails from "./Pages/admin/CategoryDetails/CategoryDetails";
import Orders from "./Pages/admin/Orders/Orders";
import Users from "./Pages/admin/Users/Users";
import UserProfile from "./Pages/admin/UserProfile/UserProfile";
import StoreSettings from "./Pages/admin/StoreSettings/StoreSettings";
import VerifyEmail from "./Pages/account/VerifyEmail";
import ResetPassword from "./Pages/account/ResetPassword";

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
      <Routes>
        <Route path="/" element={<LoadingPage />} />
        <Route path="/client/*" element={<ClientLayout />} />

        <Route path="/admin" element={<AdminLayout />}>
          <Route index element={<Dashboard />} />
          <Route path="products" element={<Products />} />
          <Route path="categories" element={<Categories />} />
          <Route path="categories/:categoryId" element={<CategoryDetails />} />
          <Route path="orders" element={<Orders />} />
          <Route path="users" element={<Users />} />
          <Route path="users/:userId" element={<UserProfile />} />
          <Route path="settings" element={<StoreSettings />} />
          <Route path="*" element={<Navigate to="/admin" replace />} />
        </Route>

        <Route path="/signUp" element={<SignUp sign={"Up"} />} />
        <Route path="/signIn" element={<SignIn sign={"In"} />} />
        {/* Opened from emails (paths match the backend's EMAIL_VERIFY_URL / PASSWORD_RESET_URL) */}
        <Route path="/verify-email" element={<VerifyEmail />} />
        <Route path="/reset-password" element={<ResetPassword />} />
        <Route path="*" element={<Navigate to="/client/home" replace />} />
      </Routes>
    </div>
  );
}

export default App;
