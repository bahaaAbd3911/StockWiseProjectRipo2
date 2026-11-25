// Import Style
import {
  CommentsFilterContianerClass,
  activeFilterClass,
  darkCommentsFilterContianerClass,
} from "./CommentOnProducts.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";

function CommentsFilter({ filters, onSelect }) {
  const { dark, lang } = useContext(AuthContext);

  return (
    <div className={`${CommentsFilterContianerClass} ${dark && darkCommentsFilterContianerClass}`}>
      {filters.map((filter) => (
        <button
          type="button"
          key={filter.label}
          className={filter.isSelected ? activeFilterClass : ""}
          aria-pressed={filter.isSelected}
          onClick={() => onSelect(filter)}
        >
          <div>{lang === "arb" ? filter.arbName : filter.name}</div>
          <div>({filter.count})</div>
        </button>
      ))}
    </div>
  );
}

export default CommentsFilter;
