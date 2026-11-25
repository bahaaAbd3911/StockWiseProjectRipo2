import styles from "./ui.module.css";

// variant: primary | outline | ghost | danger
export default function Button({
  variant = "primary",
  icon,
  children,
  type = "button",
  ...rest
}) {
  return (
    <button type={type} className={`${styles.btn} ${styles[variant]}`} {...rest}>
      {icon && <i className={icon} />}
      {children}
    </button>
  );
}
