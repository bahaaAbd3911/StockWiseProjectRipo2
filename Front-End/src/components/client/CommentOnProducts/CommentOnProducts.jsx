import { useState } from "react";
import { motion } from "motion/react";

import {
  CommentOnProductsContainerClass,
  commentStaticsAddCommentContainerClass,
  commentsContainerClass,
  commentsSectionContainerClass,
  arbCommentOnProductsContainerClass,
} from "./CommentOnProducts.module.css";

import { useLang } from "../../../hooks/useLang";
import { useShop } from "../../../hooks/useShop";
import { useProductReviews } from "../../../hooks/useProductReviews";
import { calcCountForEachRate } from "../../../utils/helpers";

import ProductStatics from "../ProductStatics/ProductStatics";
import AddComment from "../AddComment/AddComment";
import Comment from "./Comment";
import AddCommentForm from "./AddCommentForm";
import CommentsFilter from "./CommentsFilter";

const FILTERS = [
  { label: "All", name: "All Reviews", arbName: "كل الآراء" },
  ...[5, 4, 3, 2, 1].map((n) => ({ label: n, name: `${n} Stars`, arbName: `${n} نجوم` })),
];

function CommentOnProducts({ productID }) {
  const { tr, isAr } = useLang();
  const { requireSignIn } = useShop();
  const { reviews, add } = useProductReviews(productID);
  const [showForm, setShowForm] = useState(false);
  const [selected, setSelected] = useState("All");

  const counts = calcCountForEachRate(reviews);
  const filters = FILTERS.map((f) => ({
    ...f,
    isSelected: f.label === selected,
    count: f.label === "All" ? reviews.length : counts[f.label - 1],
  }));
  const visible = selected === "All" ? reviews : reviews.filter((r) => r.rating === selected);

  return (
    <motion.div
      className={`${CommentOnProductsContainerClass} ${isAr ? arbCommentOnProductsContainerClass : ""}`}
      initial={{ opacity: 0, y: 40 }}
      whileInView={{ opacity: 1, y: 0 }}
      viewport={{ once: true, amount: 0.1 }}
      transition={{ duration: 0.4 }}
    >
      <h1>
        <span>
          <i className="fa-regular fa-message"></i>
        </span>
        <span>{tr("Customer Reviews", "آراء المستخدمين")}</span>
      </h1>
      <h4>{tr("See what other customers say about this product", "ماذا يقول المستخدمون الآخرون عن هذا المنتج")}</h4>

      <div className={commentStaticsAddCommentContainerClass}>
        {showForm && <AddCommentForm onSubmit={add} onClose={() => setShowForm(false)} />}
        <ProductStatics reviews={reviews} />
        <AddComment onOpen={() => requireSignIn() && setShowForm(true)} />
      </div>

      <div className={commentsSectionContainerClass}>
        <CommentsFilter filters={filters} onSelect={(f) => setSelected(f.label)} />
        <div className={commentsContainerClass}>
          {visible.length === 0 && <p>{tr("No reviews yet — be the first!", "لا توجد آراء بعد — كن أول من يكتب!")}</p>}
          {visible.map((comment) => (
            <Comment key={comment.id} comment={comment} />
          ))}
        </div>
      </div>
    </motion.div>
  );
}

export default CommentOnProducts;
