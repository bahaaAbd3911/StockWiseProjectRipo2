import { useState } from "react";

import { AuthContext } from "./AuthContext";
import { useSession } from "../hooks/useSession";
import { EMPTY_SEARCH_FILTERS, ORDER_STAGES, PROFILE_TABS, emptyOrderForm } from "../data/clientUi";

const readTheme = () => {
  try {
    return JSON.parse(localStorage.getItem("theme")) || false;
  } catch {
    return false;
  }
};

// Session + UI preferences shared by the whole app.
export const AuthProvider = ({ children }) => {
  const session = useSession();

  const [dark, setDark] = useState(readTheme);
  const [lang, setLang] = useState(() => localStorage.getItem("lang") || "eng");
  const [searchFilters, setSearchFilters] = useState(EMPTY_SEARCH_FILTERS);
  const [message, setMessage] = useState({ isShow: false, msgTxt: "", msgBGColor: "" });
  const [userList, setUserList] = useState(PROFILE_TABS);
  const [orderStages, setOrderStages] = useState(ORDER_STAGES);
  const [orderFormFields, setorderFormFields] = useState(emptyOrderForm);

  // Opens a tab of the client profile page (0 profile, 1 orders, 2 cart, 3 favorites).
  const openProfileTab = (index) =>
    setUserList((prev) => prev.map((tab) => ({ ...tab, isActive: tab.index === index })));

  return (
    <AuthContext.Provider
      value={{
        ...session,
        dark,
        setDark,
        lang,
        setLang,
        searchFilters,
        setSearchFilters,
        message,
        setMessage,
        userList,
        setUserList,
        openProfileTab,
        orderStages,
        setOrderStages,
        orderFormFields,
        setorderFormFields,
        resetOrderForm: () => setorderFormFields(emptyOrderForm()),
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};
