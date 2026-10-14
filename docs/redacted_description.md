# Getting `redacted_description`

`redacted_description` is the lead text Facebook shows at the top of a Marketplace
listing. It is the most important field the extractor collects, and it is **not**
available from the search feed — it only comes from a separate Product Detail Page
(PDP) GraphQL request.

This document lists the requirements to obtain it. The same requirements are also
documented in code:

- `src/product.go` — `ProductDetailsGet` (extraction + full requirements)
- `src/pages.go` — `OpenFirstProduct` (why we must click a listing)
- `src/context_manager.go` — `OnResponse` (where the response is intercepted)

## Where it lives

In the PDP GraphQL response:

```
data.viewer.marketplace_product_details_page.target.redacted_description.text
```

Extracted in `ProductDetailsGet` (`src/product.go`) into `MarketplaceItemDetails.Description`.

## Requirements

1. **An authenticated Facebook session.**
   The PDP request is only meaningful when logged in; otherwise Facebook returns a
   login wall. The scraper launches Chrome with an existing logged-in profile via
   `FME_USER_DATA_DIR` / `FME_PROFILE_DIRECTORY` (see `.env`). A brand new,
   cookie-less browser profile triggers a captcha/security check instead.

2. **A real PDP request must be triggered by opening a listing.**
   Facebook only sends the `MarketplacePDPContainerQuery` request when a listing is
   opened through a **client-side click** on a search result. A plain URL navigation
   (`Page.Goto`) does **not** reliably trigger it. That is why, after the search,
   `Pages.OpenFirstProduct` clicks the first listing link.

3. **The response must be intercepted.**
   `ContextEventHandlers.OnResponse` (`src/context_manager.go`) listens for requests
   to `https://www.facebook.com/api/graphql/` whose friendly name
   (`fb_api_req_friendly_name` in the request post-data) is not in the skip set, and
   writes the decoded body to:

   ```
   data/MarketplacePDPContainerQuery/response_<timestamp>_<rand>.json
   ```

4. **The field is extracted and merged.**
   `ProcessData` runs the extractors over the captured responses. `ProductDetailsGet`
   reads `target.redacted_description.text`; the result is merged into the listing
   record at `data/__Details/detail_<id>.json` (the `FileStore` only overwrites
   non-nil fields, so it adds the description to the entry created from the search feed).

## Flow (search action)

1. Log in with the seeded Chrome profile (`FME_USER_DATA_DIR`).
2. `MarketpaceSearch` navigates to the Marketplace search URL and scrolls
   (bounded by `-scroll-seconds`).
3. Search responses are captured as `CometMarketplaceSearchContentPaginationQuery`.
4. `Pages.OpenFirstProduct` clicks the first listing → triggers
   `MarketplacePDPContainerQuery`.
5. `OnResponse` captures that PDP response containing `redacted_description`.
6. `ProcessData` writes/merges it into `data/__Details/detail_<id>.json`.

## Notes / limitations

- The PDP request that is triggered by the click also provides the reusable headers
  and post-data (`targetId`) that `RunRequestDecompress` replays to fetch details for
  the other listings.
- Only the item that is actually **opened** is guaranteed to return a full
  description. The bulk replay path currently returns small/empty responses for the
  other listings, so those may be missing a description until that path is fixed.
