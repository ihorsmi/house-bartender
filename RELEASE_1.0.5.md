# House Bartender 1.0.5

Release date: 2026-03-23

## Highlights

- Full cross-portal redesign across login, guest, bartender, and admin experiences.
- New guest ordering flow centered on a dedicated cocktail detail page and clearer order-history tracking.
- Faster bartender service flow with dashboard previews, search-first inventory, and one-click order completion.
- Cleaner admin management surfaces with aligned `Users` and `System Control` screens.

## Compared to 1.0.4

### Cross-portal redesign

- Rebuilt the shared shell around the exported Synthetica language instead of leaving each screen to feel like a separate subsystem.
- Standardized page titles, header rhythm, shared search placement, and primary action styling across the portals.
- Fixed flash notifications so login, duty, and other system notices dismiss correctly instead of lingering on screen.

### User portal

- Reworked `The Library` into a cleaner live catalog with spirit chips and shared search.
- Added a dedicated cocktail detail page with hero imagery, ingredient availability, and the full guest order form.
- Simplified `Order History` so guests can follow bartender assignment and status transitions more easily.

### Bartender portal

- Reworked `Service Dashboard` with live counts, queue preview, and direct operations shortcuts.
- Simplified `Live Queue` so bartenders can clear an order with a single `Complete Order` action while still preserving intermediate status events in history.
- Rebuilt `Inventory` into a search-first stock room with inline stock controls and a cleaner ingredient editor.
- Refreshed the cocktail library and recipe editor to match the same visual system and ingredient-rule model.

### Admin portal

- Rebuilt `Users` into a cleaner directory/editor split with role, password, access, and duty controls in one place.
- Reworked `System Control` to present counts, paths, and supported maintenance actions in the same shell.
- Kept the existing admin POST routes and management behavior intact while modernizing the presentation layer.

### Documentation refresh

- Replaced the older screenshot references with the current `docs/screenshots` set for the redesigned UI.
- Added a new release note set that documents the refreshed screens and the service-flow changes from `1.0.4`.

## Screen tour

### Login

The login screen is now cleaner and more direct, with a single `Authorize Access` action and a presentation layer that matches the rest of the release.

![Login form](docs/screenshots/loginform.png)

### User library

`The Library` now surfaces live cocktails in a cleaner catalog, with `All Spirits`, `Whiskey`, `Gin`, `Tequila`, and `Rum` chips working alongside shared search.

![User library](docs/screenshots/userlibrary.png)

### Cocktail detail and order form

Guests now place orders from a dedicated detail page that keeps availability, ingredients, quantity, location, notes, and the `Order Now` action in one flow.

![User order form](docs/screenshots/userorderform.png)

### User order history

The guest queue view now makes bartender pickup and status transitions much easier to follow, with a more readable timeline and cleaner status cards.

![User order history](docs/screenshots/userorderhistory.png)

### Service dashboard

The bartender landing screen now acts as a true operations hub, combining live counts, order preview, performance indicators, and shortcuts into the main work areas.

![Bartender dashboard](docs/screenshots/bartenderdashboard.png)

### Live queue

The queue screen now focuses on clearing tickets faster. `Complete Order` is the key flow change here, reducing bartender steps without losing the order-event trail.

![Bartender live queue](docs/screenshots/bartenderorders.png)

### Inventory

Inventory now uses a search-first header and inline stock controls so the primary work stays close to the table, while the `Add Ingredient` action remains easy to reach.

![Bartender inventory](docs/screenshots/bartenderingridients.png)

### Cocktail editor

The recipe builder now presents ingredient rows, `Required` versus `Optional` rules, media, instructions, and final save actions in a clearer publishing flow.

![Bartender cocktail editor](docs/screenshots/bartenderaddcoctails.png)

### Admin users

The admin directory now keeps `Save Details`, password rotation, access toggles, and bartender duty controls in one consistent management surface.

![Admin users](docs/screenshots/adminportal-users.png)

## Operational notes

- No manual seed step is required to move from `1.0.4` to `1.0.5`. Start the app normally and let migrations run.
- SSE remains the live-update mechanism for queue and order refreshes.
- If you linked screenshots in outside docs, update those links to the new filenames in `docs/screenshots/`.
