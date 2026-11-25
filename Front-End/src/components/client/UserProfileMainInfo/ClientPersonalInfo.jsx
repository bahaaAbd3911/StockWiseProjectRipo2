import { useContext } from "react";

import {
  ClientPersonalInfoContianerClass,
  ClientPersonalInfoHeaderClass,
  ClientPersonalInfoDataClass,
  ClientPersonalInfoEditBtnClass,
} from "./UserProfileMainInfo.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useLang } from "../../../hooks/useLang";

function ClientPersonalInfo({ onEdit }) {
  const { user } = useContext(AuthContext);
  const { tr } = useLang();

  const rows = [
    ["fa-solid fa-user", tr("Full Name", "الاسم الكامل"), user.userFullName],
    ["fa-solid fa-envelope", tr("Email Address", "البريد الإلكتروني"), user.userEmail],
    ["fa-solid fa-phone", tr("Phone Number", "رقم الهاتف"), user.phone || "—"],
  ];

  return (
    <div className={ClientPersonalInfoContianerClass}>
      <div className={ClientPersonalInfoHeaderClass}>
        <div>
          <i className="fa-solid fa-user"></i>
        </div>
        <div>{tr("Personal Information", "المعلومات الشخصية")}</div>
      </div>
      <div className={ClientPersonalInfoDataClass}>
        {rows.map(([icon, label, value]) => (
          <div key={icon}>
            <div>
              <div>
                <i className={icon}></i>
              </div>
              <div>{label}</div>
            </div>
            <div style={{ wordBreak: "break-word" }}>{value}</div>
          </div>
        ))}
      </div>
      <button type="button" className={ClientPersonalInfoEditBtnClass} onClick={onEdit}>
        <div>
          <i className="fa-solid fa-pen"></i>
        </div>
        <div>{tr("Edit Information", "تعديل المعلومات")}</div>
      </button>
    </div>
  );
}

export default ClientPersonalInfo;
