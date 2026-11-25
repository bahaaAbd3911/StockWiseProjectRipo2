// Import Style
import {
  userSettingsContainerClass,
  DarkModeButtonClass,
  chooseLanguageClass,
  activeLangClass,
  userSettingsContainerArbClass,
  settingArbClass,
  darkBgClass,
  signSettingsClass,
  startScrollClass,
  listContainerClass,
  chooseLanguageContainerClass,
  activeModeClass,
} from "./UserSettings.module.css";

// _chooseLanguageContainerClass_h23q3_309

// Import Hooks
import { useState, useContext, useEffect } from "react";
import { AuthContext } from "../../../context/AuthContext";
import { useNavigate } from "react-router";


function UserSettings() {
  const [scrollPosition, setScrollPosition] = useState(0);
  const [isOpen, setIsOpen] = useState(false);
  let navigate = useNavigate();

  useEffect(() => {
    const handleScroll = () => {
      setScrollPosition(window.scrollY);
    };

    window.addEventListener("scroll", handleScroll);

    return () => {
      window.removeEventListener("scroll", handleScroll);
    };
  }, []);
  const { dark, setDark, lang, setLang, isSign, isAdmin, user, signOut, openProfileTab } =
    useContext(AuthContext);
  return (
    <>
      <div
        className={`${userSettingsContainerClass} ${lang === "arb" ? userSettingsContainerArbClass : ""} ${dark ? darkBgClass : ""} ${scrollPosition > 100 ? startScrollClass : ""}`}
      >
        <div className={lang === "arb" ? settingArbClass : ""}>
          <div className={signSettingsClass}>
            {isSign ? (
              <>
                <div>
                  <i className="fa-solid fa-user"></i>
                </div>
                <div>{user?.userFullName}</div>
                <div
                  onClick={() => {
                    setIsOpen((prev) => !prev);
                  }}
                >
                  {!isOpen && <i className="fa-solid fa-angle-down"></i>}
                  {isOpen && <i className="fa-solid fa-angle-up"></i>}
                </div>
                {isOpen && (
                  <>
                    <div className={listContainerClass}>
                      {isAdmin && (
                        <div
                          onClick={() => {
                            setIsOpen(false);
                            navigate("/admin");
                          }}
                        >
                          <span>{lang === "arb" ? "لوحة التحكم" : "Admin Panel"}</span>
                          <span>
                            <i className="fa-solid fa-gauge"></i>
                          </span>
                        </div>
                      )}
                      <div
                        onClick={() => {
                          setIsOpen(false);
                          openProfileTab(0);
                          navigate("/client/profile");
                        }}
                      >
                        <span>
                          {lang === "arb" ? "الصفحة الشخصية" : "Profile"}
                        </span>
                        <span>
                          <i className="fa-solid fa-user"></i>
                        </span>
                      </div>
                      <div
                        onClick={() => {
                          setIsOpen(false);
                          signOut();
                          navigate("/client/home");
                        }}
                      >
                        <span>
                          {lang === "arb" ? "تسجيل الخروج" : "Sign Out"}
                        </span>
                        <span>
                          <i className="fa-solid fa-right-from-bracket"></i>
                        </span>
                      </div>
                    </div>
                  </>
                )}
              </>
            ) : (
              <>
                <div
                  onClick={() => {
                    navigate("/signIn");
                  }}
                >
                  {lang === "arb" ? "تسجيل الدخول" : "Sign In"}
                </div>
                <div
                  onClick={() => {
                    navigate("/signUp");
                  }}
                >
                  {lang === "arb" ? "انشاء حساب" : "Sign Up"}
                </div>
              </>
            )}
          </div>
        </div>
        <div className={lang === "arb" ? settingArbClass : ""}>
          <div
            className={DarkModeButtonClass}
            onClick={() => {
              localStorage.setItem(
                "theme",
                !JSON.parse(localStorage.getItem("theme")),
              );
              setDark(JSON.parse(localStorage.getItem("theme")));
            }}
          >
            <div className={dark === false ? activeModeClass : ""}>
              <i className="fa-solid fa-sun"></i>
            </div>
            <div className={dark === true ? activeModeClass : ""}>
              <i className="fa-solid fa-moon"></i>
            </div>
          </div>
        </div>
        <div
          className={`${chooseLanguageContainerClass} ${lang === "arb" ? settingArbClass : ""}`}
        >
          <div
            className={`${chooseLanguageClass} ${lang === "arb" ? activeLangClass : ""}`}
            onClick={() => {
              localStorage.setItem("lang", "arb");
              setLang(localStorage.getItem("lang"));
            }}
          >
            عربي
          </div>
          <div
            className={`${chooseLanguageClass} ${lang === "eng" ? activeLangClass : ""}`}
            onClick={() => {
              localStorage.setItem("lang", "eng");
              setLang(localStorage.getItem("lang"));
            }}
          >
            Eng
          </div>
        </div>
      </div>
    </>
  );
}
export default UserSettings;
