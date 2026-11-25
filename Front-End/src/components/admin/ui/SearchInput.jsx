import styles from "./ui.module.css";

export default function SearchInput({ value, onChange, placeholder }) {
  return (
    <div className={styles.searchBox}>
      <i className="fa-solid fa-magnifying-glass" />
      <input
        type="search"
        className={styles.input}
        value={value}
        placeholder={placeholder}
        onChange={(e) => onChange(e.target.value)}
      />
    </div>
  );
}
