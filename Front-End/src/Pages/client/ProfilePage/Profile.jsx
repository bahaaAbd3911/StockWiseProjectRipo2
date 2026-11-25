import { useState } from "react";
import { motion } from "motion/react";

import {
  ProfileContainerClass,
  ordersAndInfoContainerClass,
  arbProfileContainerClass,
} from "./ClientProfilePage.module.css";
import { useLang } from "../../../hooks/useLang";

import UserProfileMainInfo from "../../../components/client/UserProfileMainInfo/UserProfileMainInfo";
import ClitentStatics from "../../../components/client/UserProfileMainInfo/ClitentStatics";
import ClientPersonalInfo from "../../../components/client/UserProfileMainInfo/ClientPersonalInfo";
import RecentOrders from "../../../components/client/UserProfileMainInfo/RecentOrders";
import ProfileEditModal from "./ProfileEditModal";

// "My Profile" tab: the signed-in user's real data, counters and recent orders.
function Profile() {
  const { isAr } = useLang();
  const [editing, setEditing] = useState(false);
  const openEditor = () => setEditing(true);

  return (
    <div className={`${ProfileContainerClass} ${isAr ? arbProfileContainerClass : ""}`}>
      <UserProfileMainInfo onEdit={openEditor} />
      <ClitentStatics />
      <motion.div
        className={ordersAndInfoContainerClass}
        initial={{ opacity: 0, y: 30 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.35 }}
      >
        <ClientPersonalInfo onEdit={openEditor} />
        <RecentOrders />
      </motion.div>
      {editing && <ProfileEditModal onClose={() => setEditing(false)} />}
    </div>
  );
}

export default Profile;
