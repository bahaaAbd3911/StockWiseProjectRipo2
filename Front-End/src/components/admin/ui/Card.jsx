import styles from "./ui.module.css";

export default function Card({ className = "", children, ...rest }) {
  return (
    <section className={`${styles.card} ${className}`} {...rest}>
      {children}
    </section>
  );
}
