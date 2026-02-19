# House Bartender 1.0.1

Release date: 2026-02-19

## Highlights

- Fixed cocktail create/edit usability issues in the bartender portal.
- Fixed boolean form parsing so enabled/available toggles persist correctly.
- Added seed catalog support for single-serve beverages mapped from products.

## Changes

### Fixed

- Fixed `New Cocktail` form rendering path by wiring `cocktail_form.html` in the main layout template.
- Fixed ingredient dropdown row binding in cocktail form so product options and submit controls render reliably.
- Fixed cocktail table actions:
  - corrected toggle payload field to `is_enabled`
  - corrected edit link route to `/bartender/cocktails/{id}/edit`
- Fixed checkbox+hidden boolean parsing bug that caused `is_enabled` / `is_available` to be read as false in some submissions.

### Added

- Added shared form boolean parser used by handlers to correctly interpret checkbox values.
- Added seeded single-serve cocktails and ingredient mappings:
  - `Water`
  - `Red Wine`
  - `Beer (Lager)`

## Operational notes

- Run catalog seed from Admin Settings after upgrade to populate newly seeded cocktails on existing databases.
