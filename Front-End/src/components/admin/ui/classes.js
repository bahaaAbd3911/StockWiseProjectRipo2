import styles from "./ui.module.css";

// Shared class names for places that need raw elements (tables, native inputs).
export const tableClasses = {
  wrap: styles.tableWrap,
  table: styles.table,
  clickableRow: styles.clickableRow,
  rowActions: styles.rowActions,
};

export const toolbarClass = styles.toolbar;
export const formGridClass = styles.formGrid;
export const inputClass = styles.input;
export const selectClass = styles.select;
