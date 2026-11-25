import styles from "./ui.module.css";

// Label + input (or a custom control passed as children) + error/hint text.
export default function FormField({
  label,
  error,
  hint,
  as: Tag = "input",
  children,
  id,
  ...inputProps
}) {
  const fieldId = id ?? inputProps.name;

  return (
    <div className={styles.field}>
      {label && <label htmlFor={fieldId}>{label}</label>}

      {children ?? (
        <Tag
          id={fieldId}
          className={`${styles.input} ${error ? styles.inputError : ""}`}
          {...inputProps}
        />
      )}

      {error ? (
        <span className={styles.fieldError}>{error}</span>
      ) : (
        hint && <span className={styles.fieldHint}>{hint}</span>
      )}
    </div>
  );
}
