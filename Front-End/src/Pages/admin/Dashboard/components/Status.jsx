import { FaCheckCircle, FaClock, FaTruck, FaTimesCircle, FaCog } from "react-icons/fa";
import styles from "../Dashboard.module.css";

// Order status -> [icon, style class]
const STATUS = {
  pending: [FaClock, "pending"],
  confirmed: [FaCheckCircle, "pending"],
  processing: [FaCog, "shipped"],
  shipped: [FaTruck, "shipped"],
  delivered: [FaCheckCircle, "completed"],
  cancelled: [FaTimesCircle, "cancelled"],
};

export default function Status({ status, t }) {
  const [Icon, className] = STATUS[status] ?? STATUS.pending;

  return (
    <span className={`${styles.status} ${styles[className]}`}>
      <Icon /> {t[status]}
    </span>
  );
}
