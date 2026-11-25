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
import LoadingPage from "./pages/LoadingPage/LoadingPage";
import Dashboard from "./pages/admin/Dashboard/Dashboard";
import Products from "./pages/admin/Products/Products";
import Categories from "./pages/admin/Categories/Categories";
import CategoryDetails from "./pages/admin/CategoryDetails/CategoryDetails";
import Orders from "./pages/admin/Orders/Orders";
import Users from "./pages/admin/Users/Users";
import UserProfile from "./pages/admin/UserProfile/UserProfile";
import StoreSettings from "./pages/admin/StoreSettings/StoreSettings";

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
        <Route path="*" element={<Navigate to="/client/home" replace />} />
      </Routes>
    </div>
  );
}

export default App;
