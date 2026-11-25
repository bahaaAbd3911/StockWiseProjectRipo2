import styles from "./ui.module.css";

// options: [{ value, label }]
export default function SelectInput({ value, onChange, options, ...rest }) {
  return (
    <select
      className={styles.select}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      {...rest}
    >
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
}
