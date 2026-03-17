# House Bartender 1.0.3

Release date: 2026-03-17

## Highlights

- Standardized more portal screens around the same stacked, service-friendly layout.
- Fixed inline inventory updates so stock and manual availability can be managed reliably from the list view.
- Made compact line view the default cocktail layout, with optional grid view switching for bartender and user portals.

## Changes

### Changed

- Realigned Admin Users so `Create user` appears above `All users`, matching the inventory-style stacked layout.
- Realigned the bartender dashboard so `Queue snapshot` appears above `Newest orders` with matching panel/header spacing.
- Added cocktail view switching on bartender and user catalog screens.
- Made the compact line view the default cocktail layout and kept grid view as an optional toggle.

### Fixed

- Fixed custom HTMX form behavior so inline POST forms submit on `submit` instead of firing from input clicks.
- Fixed inline inventory form requests to send URL-encoded values that Go handlers parse correctly.
- Fixed inventory stock updates from the list screen.
- Fixed clearing `stock_count` back to blank so products can return to manual availability control.
- Fixed `Mark available` / `Mark unavailable` actions from the inventory list view.
- Fixed HTMX inventory actions to refresh the products table instead of forcing a full page redirect.

## Operational notes

- No data migration or seed action is required for this release.
- After upgrade, bartender and user cocktail screens default to line view. Grid view remains available from the toggle button.
