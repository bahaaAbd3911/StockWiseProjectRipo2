import { useEffect, useState } from "react";

import { fetchReviews, postReview } from "../api/catalog";
import { useCatalog } from "./useCatalog";

// Reviews of one product. add() posts a review and refreshes the product's rating.
export function useProductReviews(productId) {
  const { reload } = useCatalog();
  const [reviews, setReviews] = useState([]);

  useEffect(() => {
    fetchReviews(productId).then(setReviews).catch(() => setReviews([]));
  }, [productId]);

  const add = async (comment, rating) => {
    const review = await postReview(productId, comment, rating);
    setReviews((prev) => [review, ...prev]);
    reload();
    return review;
  };

  return { reviews, add };
}

export const averageRating = (reviews) =>
  reviews.length ? reviews.reduce((sum, r) => sum + r.rating, 0) / reviews.length : 0;
