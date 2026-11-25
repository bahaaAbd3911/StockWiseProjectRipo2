import { useLang } from "../../../hooks/useLang";
import {
  chooseProductColorContainerClass,
  chooseColorClass,
  choosenColorClass,
  activeColorClass,
  sectionLabelClass,
} from "./ProductDetails.module.css";

function ChooseProductColor({ productData, color, setColor }) {
  const { tr, isAr } = useLang();
  const colors = productData.availableColors ?? [];
  if (colors.length === 0) return null;

  const selected = colors.find((c) => c.name === color);

  return (
    <div>
      <p className={sectionLabelClass}>
        {tr("Color", "اللون")}: {selected ? (isAr ? selected.nameAr : selected.name) : ""}
      </p>
      <div className={chooseProductColorContainerClass} role="radiogroup">
        {colors.map((c) => (
          <button
            type="button"
            key={c.name}
            role="radio"
            aria-checked={color === c.name}
            aria-label={isAr ? c.nameAr : c.name}
            title={isAr ? c.nameAr : c.name}
            className={`${choosenColorClass} ${color === c.name ? activeColorClass : ""}`}
            style={{ background: "transparent" }}
            onClick={() => setColor(c.name)}
          >
            <div className={chooseColorClass} style={{ background: c.hex }}></div>
          </button>
        ))}
      </div>
    </div>
  );
}

export default ChooseProductColor;
