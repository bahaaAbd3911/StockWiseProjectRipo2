import { useContext } from "react";
import { motion } from "motion/react";

import {
  UserProfileMainInfoContainerClass,
  userImgClass,
  userDataContainerClass,
  eidtBtnContainerClass,
  starBgClass,
} from "./UserProfileMainInfo.module.css";
import { AuthContext } from "../../../context/AuthContext";
import { useLang } from "../../../hooks/useLang";
import { formatDate, initials } from "../../../utils/format";

function UserProfileMainInfo({ onEdit }) {
  const { user } = useContext(AuthContext);
  const { tr, isAr } = useLang();

  return (
    <motion.div
      className={UserProfileMainInfoContainerClass}
      initial={{ opacity: 0, y: -40 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.35 }}
    >
      <div>
        <div className={userImgClass}>
          {user.image ? (
            <img src={user.image} alt="" />
          ) : (
            <span
              style={{
                width: "100%",
                height: "100%",
                display: "grid",
                placeItems: "center",
                borderRadius: "50%",
                fontSize: 32,
                fontWeight: 800,
                background: "var(--color-primary)",
                color: "var(--color-on-primary)",
              }}
            >
              {initials(user.userFullName)}
            </span>
          )}
        </div>
        <div className={userDataContainerClass}>
          <div>
            <div>{user.userFullName}</div>
            <div className={starBgClass}>
              <div className="fa-solid fa-certificate">
                <div className="fa-solid fa-check"></div>
              </div>
            </div>
          </div>
          <div>
            <div>
              <i className="fa-solid fa-envelope"></i>
            </div>
            <div>{user.userEmail}</div>
          </div>
          <div>
            <div>
              <i className="fa-solid fa-calendar-days"></i>
            </div>
            <div>
              {tr("Member since", "عضو منذ")} {formatDate(user.joinedAt, isAr)}
            </div>
          </div>
        </div>
      </div>
      <button type="button" className={eidtBtnContainerClass} onClick={onEdit}>
        <div>
          <i className="fa-solid fa-pen"></i>
        </div>
        <div>{tr("Edit Profile", "تعديل المعلومات")}</div>
      </button>
    </motion.div>
  );
}

export default UserProfileMainInfo;
