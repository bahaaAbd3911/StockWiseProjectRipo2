import { useState } from "react";

import {
  FormInputContainerClass,
  listContainerClass,
  choosenItemClass,
  errorMsgClass,
  beforeLastClass,
} from "./ShippingForm.module.css";

// Click-to-open list used for City and Area. options: [{ value, label }]
export default function ChoicePicker({ label, placeholder, value, options, onSelect, error, disabled }) {
  const [open, setOpen] = useState(false);
  const selected = options.find((o) => o.value === value);

  return (
    <div className={FormInputContainerClass}>
      <div>
        <label>{label}</label>
        <span>*</span>
      </div>
      <div
        role="button"
        tabIndex={0}
        aria-expanded={open}
        aria-disabled={disabled}
        className={`${choosenItemClass} ${beforeLastClass}`}
        style={disabled ? { opacity: 0.55, cursor: "not-allowed" } : undefined}
        onClick={() => !disabled && setOpen((v) => !v)}
        onKeyDown={(e) => e.key === "Enter" && !disabled && setOpen((v) => !v)}
      >
        <div>{selected?.label ?? placeholder}</div>
        <div>
          <i className={`fa-solid ${open ? "fa-angle-up" : "fa-chevron-down"}`}></i>
        </div>
      </div>
      {open && (
        <div className={listContainerClass} role="listbox">
          {options.map((option) => (
            <div
              key={option.value}
              role="option"
              aria-selected={option.value === value}
              onClick={() => {
                onSelect(option.value);
                setOpen(false);
              }}
            >
              {option.label}
            </div>
          ))}
        </div>
      )}
      <span className={errorMsgClass}>{error}</span>
    </div>
  );
}
