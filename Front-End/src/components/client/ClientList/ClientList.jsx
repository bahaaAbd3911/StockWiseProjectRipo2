import { useContext } from "react";
import { useNavigate } from "react-router";

import {
  ClientListContainerClass,
  itemContaienrClass,
  activeItemClass,
  arbClientListContainerClass,
} from "./ClientList.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useShop } from "../../../hooks/useShop";

const LOG_OUT = 4;

// Profile page tabs. Shows live counts for orders, cart and favorites.
function ClientList() {
  const { userList, openProfileTab, lang, signOut } = useContext(AuthContext);
  const { orders, cartCount, favorites } = useShop();
  const navigate = useNavigate();
  const counts = { 1: orders.length, 2: cartCount, 3: favorites.length };

  const select = (index) => {
    if (index === LOG_OUT) {
      signOut();
      openProfileTab(0);
      navigate("/client/home");
      return;
    }
    openProfileTab(index);
  };

  return (
    <nav className={`${ClientListContainerClass} ${lang === "arb" ? arbClientListContainerClass : ""}`}>
      {userList.map((item) => (
        <button
          type="button"
          key={item.index}
          className={`${itemContaienrClass} ${item.isActive ? activeItemClass : ""}`}
          aria-current={item.isActive ? "page" : undefined}
          onClick={() => select(item.index)}
        >
          <div>
            <i className={item.icon}></i>
          </div>
          <div>{lang === "arb" ? item.arbLabel : item.label}</div>
          {counts[item.index] > 0 && <span style={{ marginInlineStart: "auto", fontWeight: 700 }}>{counts[item.index]}</span>}
        </button>
      ))}
    </nav>
  );
}

export default ClientList;
