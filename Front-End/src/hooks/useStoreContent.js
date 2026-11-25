import { useContext } from "react";

import { StoreContext } from "../context/StoreContext";
import {
  DEFAULT_FEATURES,
  DEFAULT_FOOTER,
  DEFAULT_HERO,
  DEFAULT_PRODUCTS_BANNER,
  withBannerDefaults,
} from "../data/storeContent";

// The current store's page content with the defaults filled in. Settings come
// from the API for the store this address belongs to, so the same components
// show each store's own banners, tiles and footer.
export function useStoreContent() {
  const { storeSettings } = useContext(StoreContext);
  const content = storeSettings?.content ?? {};
  const footer = content.footer ?? {};

  return {
    hero: withBannerDefaults(content.hero, DEFAULT_HERO),
    productsBanner: withBannerDefaults(content.productsBanner, DEFAULT_PRODUCTS_BANNER),
    features: content.features?.length ? content.features : DEFAULT_FEATURES,
    footer: {
      ...DEFAULT_FOOTER,
      ...Object.fromEntries(Object.entries(footer).filter(([, v]) => typeof v !== "string" || v.trim())),
      social: footer.social ?? {},
    },
  };
}

// A banner in the shape the Advertisement component uses.
export const toAdData = (b) => ({
  adTitle: b.title,
  arbAdTitle: b.titleAr,
  adSubTitle: b.subtitle,
  arbAdSubTitle: b.subtitleAr,
  com: b.highlight,
  arbCom: b.highlightAr,
  buttonText: b.buttonText,
  arbButtonText: b.buttonTextAr,
  imgSrc: b.image,
});
