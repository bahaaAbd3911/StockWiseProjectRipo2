# 8. Catalog and uploads

## 8.1 Products

A product has:

- a name, description and category in English and Arabic;
- a price;
- its own discount percentage;
- a stock count;
- images and colors;
- an on/off switch (`is_active`).

Computed when it is read:

| Field | Meaning |
|---|---|
| `final_price` | what one unit costs now: the product's own discount or the best live offer, whichever is lower ([09](09-prices-offers-coupons-bundles.md)) |
| `offer` | the offer that set that price, for the badge (null if none) |
| `in_stock` | stock > 0 |
| `primary_image_url`, `image_urls` | the main image first |
| `rating_average`, `review_count` | from reviews |

Listing: `GET /api/products` accepts these query parameters, and the result
is paginated:

| Parameter | Effect |
|---|---|
| `category`, `search` | filter |
| `sort=newest`, `price_asc`, `price_desc`, `name` | order |
| `in_stock=true` | only products in stock |
| `on_offer=true` | only products with a live offer |

**Sort values are whitelisted.** Only fixed `ORDER BY` texts from a map can
reach the SQL, so `sort` cannot be used for SQL injection. Price sorting uses
the final price including offers, so the order matches what customers see.

**Images and colors are saved with the product.** `POST` and `PUT` accept:

- `images: [url, ...]`: the first one becomes the main image;
- `colors: [{name, name_ar, hex_code}]`.

**Why?** The admin form edits everything at once; saving in one request
means a half-saved product cannot exist. Lists that are left out of the
body are kept as they are.

Further rules:

- **Exactly one primary image**, enforced by a unique index and a row lock.
  Deleting the primary promotes the next image.
- **Colors** are unique per product. A product with colors requires a color
  in the cart.
- **Delete = deactivate.** Orders still point at the product.
- **Restocking emails the waiting list.** When an admin saves a product with
  stock above 0, every waiting customer is notified
  ([12](12-reviews-favorites-addresses-waiting-list.md)).

## 8.2 Categories

Categories are real records:

- a name and an Arabic name;
- a Font Awesome icon;
- an on/off switch;
- a product count.

Products keep the category's name in their `category` field, so existing
products and filters kept working when categories became records
(migration 000006 created them from the categories already used).

| Action | Effect |
|---|---|
| rename | the new name is copied to the category's products |
| switch off | hidden from customers (admins still see it to switch it back on) |
| delete | the category is removed and its products are hidden (kept for order history) |

## 8.3 Image uploads

There are three ways an image gets into the system:

- product image by URL: `POST /api/products/{id}/images`;
- product image upload: `POST /api/admin/products/{id}/images/upload`;
- profile picture: `POST /api/profile/image`; banner images:
  `POST /api/admin/settings/images`.

```
multipart upload ─► size cap (MAX_UPLOAD_MB, default 5), read from the stream, never a temp file
                 ─► images.Inspect: really JPEG / PNG / WebP?  sensible dimensions?  strip metadata
                 ─► storage.Save: random name (32 hex + extension) in the store's folder
                      local: data/uploads/<store id>/<name>  →  /uploads/<store id>/<name>
                      s3:    bucket/uploads/<store id>/<name> →  S3_PUBLIC_URL/uploads/<store id>/<name>
                 ─► save the URL on the product / user / settings
                    (if that fails, the file is deleted again)
```

**Why check the bytes?** The file name and the content type are chosen by
the client. A "picture" that is really HTML could run scripts if served from
the shop's domain. The type is decided from the file's first bytes.

**Why only read the header?** A small file can describe a giant image (a
"decompression bomb"). Dimensions are checked from the header without
decoding the pixels.

**Why strip metadata?** Phone photos contain the GPS position where they
were taken, often the customer's home. Metadata is removed:

- EXIF and GPS;
- XMP and IPTC;
- comments;
- PNG text chunks.

JPEG orientation is kept in a minimal new EXIF block, so photos are not
shown sideways.

**Why one folder per store?** Files can be managed and backed up per store.
Deleting checks ownership (`ownedKey`): a store can only delete files in its
own folder, even if it sends another store's URL.

How files are served (local storage):

- only names the API created;
- `Cache-Control: immutable`, because a name never gets new content;
- `nosniff`;
- no directory listings.

Deleting an image, or replacing or removing a profile picture or banner,
deletes the file. URLs that were typed in (not uploaded) are left alone.

**Storage choice:** local storage needs a disk that survives deploys (the
Compose file uses a Docker volume). On hosts with a temporary disk, use S3
or Cloudflare R2 (`STORAGE_DRIVER=s3`).
