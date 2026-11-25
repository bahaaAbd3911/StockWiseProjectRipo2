// Import Style
import {
  AddCommentContainerClass,
  addReviewContainerClass,
} from "./AddComment.module.css";

// Import Hooks
import { useContext } from "react";
import { AuthContext } from "../../../context/AuthContext";

function AddComment({ onOpen }) {
  const { lang } = useContext(AuthContext);

  return (
    <>
      <div className={AddCommentContainerClass}>
        <div
          className={addReviewContainerClass}
          onClick={onOpen}
        >
          <div>
            <i className="fa-solid fa-pen"></i>
          </div>
          <div>{lang === "arb" ? "إكتب رأيك فالمنتج" : "Write A Review"}</div>
        </div>
        <div>
          {lang === "arb"
            ? `شارك تجربتك مع العملاء الآخرين وساعدهم على اتخاذ القرار الصحيح`
            : `Share Your Experience With Other Customers And Help Them Make The Right
          Choice`}
        </div>
      </div>
    </>
  );
}

export default AddComment;
