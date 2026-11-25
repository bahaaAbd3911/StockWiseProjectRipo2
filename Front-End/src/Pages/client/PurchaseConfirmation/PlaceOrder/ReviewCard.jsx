import { formClass, formTitleAndIconConatinerClass, formDataClass, editBtnClass } from "./PlaceOrder.module.css";
import Icon from "../../../../components/client/Icon/Icon";

// One summary box on the review step, with an optional "Edit" button.
export default function ReviewCard({ icon, title, lines, onEdit, editLabel }) {
  return (
    <div className={formClass}>
      <div>
        <div className={formTitleAndIconConatinerClass}>
          <Icon iconSrc={icon} />
          <div>{title}</div>
        </div>
        <div className={formDataClass}>
          {lines.map(([label, value]) => (
            <div key={label}>
              {label}: <strong>{value || "—"}</strong>
            </div>
          ))}
        </div>
      </div>
      {onEdit && (
        <div className={editBtnClass}>
          <button onClick={onEdit}>
            <span>
              <i className="fa-solid fa-pen"></i>
            </span>
            <span>{editLabel}</span>
          </button>
        </div>
      )}
    </div>
  );
}
