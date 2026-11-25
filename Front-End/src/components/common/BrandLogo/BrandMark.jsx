import { useBrand } from "../../../hooks/useBrand";

// The store logo: the uploaded image if there is one, otherwise the basket icon.
export default function BrandMark({ className = "", size = 40 }) {
  const { logo, name } = useBrand();

  if (logo) {
    return (
      <img
        src={logo}
        alt={name}
        className={className}
        style={{ height: size, width: "auto", maxWidth: size * 3, objectFit: "contain" }}
      />
    );
  }

  return <i className={`fa-solid fa-basket-shopping ${className}`} />;
}
