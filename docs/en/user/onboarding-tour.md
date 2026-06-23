# Interface onboarding tour

After the **first successful sign-in**, TreePage shows a short guided tour of the interface. The welcome splash animation (logo and greeting) finishes first — the tour starts only when it completes.

## What the tour covers

| Step | Topic |
|------|--------|
| Navigation | **Spaces**, **Search**, **Account**, **Admin** (if you are an administrator) |
| Sections | Catalog of spaces, search form, **My pages**, admin sidebar |
| Document view | Document tree and **Comments** column (on a sample page) |

Use **Next** / **Back** to move through steps. **Skip** or **Done** ends the tour.

## Restart the tour

Click the **graduation cap** icon (🎓) in the top navigation bar at any time to run the tour again.

## Technical note

Completion is stored in the browser (`localStorage` key `treepage_onboarding_v1_done`). Clearing site data or using another browser will show the tour again on first visit after login.

## Related sections

- [First login](../getting-started/first-login.md)
- [Navigation](navigation.md)
