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

## Flow

Two separate commands:

### `search` — collect listings
1. Log in with the seeded Chrome profile (`FME_USER_DATA_DIR`).
2. `MarketpaceSearch` navigates to the Marketplace search URL and scrolls
   (bounded by `-scroll-seconds`).
3. Search responses are captured as `CometMarketplaceSearchContentPaginationQuery`.
4. `save`/`process_data` extracts them into `data/__Details/detail_<id>.json`
   (without descriptions).

### `pull_description` — fetch missing descriptions
1. Log in, open the Marketplace home, and `Pages.OpenFirstProduct` clicks the first
   listing → triggers `MarketplacePDPContainerQuery`. Any listing works; it only needs
   to exist so we can capture the request.
2. `ContextEventHandlers.OnResponse` captures that first PDP request as a reusable
   template (`URL`, `method`, `headers`, post-data with `variables.targetId`).
3. For every `data/__Details/detail_<id>.json` **missing** a description,
   `FetchPDP` clones the template, sets `variables.targetId` to that listing's id,
   and replays the request **inside the page** (a `fetch(...)` executed via
   `page.Evaluate` with `credentials: "include"`). It must run in the page: Facebook
   rejects the replayed GraphQL request (error `1357054`) when it is issued outside
   the page context, even with an identical body/headers.
4. `FetchPDP` writes each response to `data/MarketplacePDPContainerQuery/`, stopping
   early if Facebook returns the rate-limit error.
5. `ProcessData` runs once at the end and merges the descriptions into the detail files.

Run it with `just pull-description` (optionally `-delay-ms=N`).

Responses are written **once per product id** at
`data/MarketplacePDPContainerQuery/response_0_<id>.json` (idempotent overwrite), and
ids that already have a response on disk are skipped. Re-running therefore never
creates duplicates. If duplicates ever appear, `just dedupe-pdp` removes them
(keeping the newest per id).

## Notes

- `ReplayHeaders` drops `content-length`/`content-encoding`/`host` from the captured
  headers (they belong to the original body); `GetPostData` rebuilds a correctly
  form-encoded body.
- Pulling stops (without failing the run) on the rate-limit error, so a very large
  batch may need more than one run. Use `-delay-ms` to slow it down if needed.



