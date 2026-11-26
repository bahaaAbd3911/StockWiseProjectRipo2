import { useContext } from "react";

import {
  ClientProfilePageConatinerClass,
  componentsContainerClass,
} from "./ClientProfilePage.module.css";
import { AuthContext } from "../../../context/AuthContext";

import UserSettings from "../../../components/common/UserSettings/UserSettings";
import SearchBar from "../../../components/common/SearchBar/SearchBar";
import ClientList from "../../../components/client/ClientList/ClientList";
import SignInRequired from "../../../components/client/SignInRequired/SignInRequired";
import ShoppingCartPage from "../ShoppingCartPage/ShoppingCartPage";
import FavoritePage from "../FavoritePage/FavoritePage";
import ClientOrdersPage from "../OrdersPage/ClientOrdersPage";
import Profile from "./Profile";
import DebtPage from "../Sales/DebtPage";
import WaitingListPage from "../Sales/WaitingListPage";

// Tabs by index: 0 profile, 1 orders, 2 cart, 3 favorites, 4 balance, 5 waiting list
// (6 = log out, handled in ClientList).
const TAB_PAGES = [Profile, ClientOrdersPage, ShoppingCartPage, FavoritePage, DebtPage, WaitingListPage];

function ClientProfilePage() {
  const { userList, isSign, ready } = useContext(AuthContext);
  const active = userList.find((item) => item.isActive)?.index ?? 0;
  const ActivePage = TAB_PAGES[active] ?? Profile;

  return (
    <div className={ClientProfilePageConatinerClass}>
      <UserSettings />
      <SearchBar value={1} />
      {!ready ? null : isSign ? (
        <div className={componentsContainerClass}>
          <ClientList />
          <ActivePage />
        </div>
      ) : (
        <SignInRequired />
      )}
    </div>
  );
}

export default ClientProfilePage;
