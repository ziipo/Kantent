# SvelteKit Frontend Implementation Plan

This document outlines the plan to create an alternative SvelteKit frontend for Kantent that mirrors all functionality of the existing React frontend.

## Current Status

The React frontend (`frontend/`) is fully functional with:
- ✅ Pinterest-style masonry grid layout
- ✅ Infinite scroll with TanStack Query
- ✅ Feed management modal with 4 tabs (Discover, RSS, Reddit, YouTube)
- ✅ YouTube channel resolution for @handle URLs
- ✅ Article modal for reading content
- ✅ Mark as read/unread and star functionality
- ✅ Feed discovery from websites
- ✅ Tailwind CSS v3 styling

A SvelteKit scaffold has been started in `frontend-svelte/` with:
- Basic configuration files (svelte.config.js, vite.config.js, tailwind.config.js)
- Dependencies installed (SvelteKit, Tailwind v3, adapter-static)
- Directory structure created

## Architecture Overview

### Directory Structure
```
frontend-svelte/
├── src/
│   ├── routes/
│   │   ├── +page.svelte          # Main page with masonry grid
│   │   ├── +layout.svelte        # Root layout with header
│   │   └── +layout.js            # Client-side only (CSR)
│   ├── lib/
│   │   ├── components/
│   │   │   ├── ArticleCard.svelte
│   │   │   ├── ArticleModal.svelte
│   │   │   ├── FeedManager.svelte
│   │   │   ├── Header.svelte
│   │   │   └── MasonryGrid.svelte
│   │   ├── api/
│   │   │   └── client.js         # API functions (same as React)
│   │   └── stores/
│   │       ├── articles.js       # Article state management
│   │       └── feeds.js          # Feed state management
│   └── app.css                   # Tailwind imports
├── static/
│   └── favicon.png
├── build/                        # Output directory
├── svelte.config.js
├── vite.config.js
├── tailwind.config.js
└── package.json
```

## Implementation Steps

### Phase 1: Core Setup (30 minutes)

1. **Create base layout and routing**
   - `src/routes/+layout.svelte` - Root layout with Tailwind
   - `src/routes/+layout.js` - Configure CSR (client-side rendering only)
   - `src/app.html` - HTML template
   - `src/app.css` - Tailwind directives

2. **Create API client**
   - `src/lib/api/client.js` - Copy from React version, identical functionality
   - All API functions: fetchArticles, createFeed, resolveYouTubeChannel, etc.

### Phase 2: State Management (30 minutes)

3. **Create Svelte stores**
   - `src/lib/stores/articles.js` - Article state with infinite scroll logic
   - `src/lib/stores/feeds.js` - Feed management state
   - `src/lib/stores/ui.js` - Modal visibility, selected article, etc.

### Phase 3: Components (2-3 hours)

4. **Header component** (`src/lib/components/Header.svelte`)
   - Logo and title
   - "Manage Feeds" button
   - Stats display (total articles, unread count)
   - Filter buttons (All, Unread)

5. **MasonryGrid component** (`src/lib/components/MasonryGrid.svelte`)
   - Use CSS grid or similar for masonry layout
   - Infinite scroll with Intersection Observer
   - Responsive breakpoints (1, 2, 3, 4 columns)
   - Loading skeleton states

6. **ArticleCard component** (`src/lib/components/ArticleCard.svelte`)
   - Image display with aspect ratio
   - Title, description, metadata
   - Read/unread indicator
   - Star button
   - Click handler for modal
   - Optimistic UI updates

7. **ArticleModal component** (`src/lib/components/ArticleModal.svelte`)
   - Full article display with content
   - Close button (X or click outside)
   - Mark read on open
   - Read/starred controls
   - Proper z-index and backdrop

8. **FeedManager component** (`src/lib/components/FeedManager.svelte`)
   - Modal with 4 tabs:
     - **Discover**: URL input, finds feeds automatically
     - **RSS**: Direct feed URL input
     - **Reddit**: Subreddit + sort options
     - **YouTube**: Channel URL with resolver
   - Feed list with delete and refresh buttons
   - Click feed to filter articles
   - Proper loading states

### Phase 4: Main Page (30 minutes)

9. **Main page** (`src/routes/+page.svelte`)
   - Import and use Header component
   - Import and use MasonryGrid component
   - Conditional rendering of ArticleModal
   - Conditional rendering of FeedManager
   - Handle all state subscriptions

### Phase 5: Styling (1 hour)

10. **Tailwind styling**
    - Ensure all classes match React version
    - Modal backdrop and animations
    - Responsive design
    - Loading states and skeletons
    - Hover effects and transitions

### Phase 6: Testing & Polish (1 hour)

11. **Testing and refinement**
    - Test all feed types (RSS, Reddit, YouTube)
    - Test YouTube @handle resolution
    - Test feed discovery
    - Test infinite scroll
    - Test mark as read/starred
    - Test modal interactions
    - Test responsive design

## Key Differences from React

### State Management
- **React**: TanStack Query with React hooks
- **Svelte**: Writable stores with reactive statements

Example:
```javascript
// React
const { data, fetchNextPage } = useInfiniteQuery(...)

// Svelte
import { articles } from '$lib/stores/articles';
$: allArticles = $articles;
```

### Infinite Scroll
- **React**: react-intersection-observer + TanStack Query
- **Svelte**: Intersection Observer API + store actions

Example:
```svelte
<script>
  import { onMount } from 'svelte';
  import { articles, loadMore } from '$lib/stores/articles';

  let observer;
  let sentinel;

  onMount(() => {
    observer = new IntersectionObserver(entries => {
      if (entries[0].isIntersecting) loadMore();
    });
    observer.observe(sentinel);

    return () => observer.disconnect();
  });
</script>

<div bind:this={sentinel}></div>
```

### Component Communication
- **React**: Props and callbacks
- **Svelte**: Props, events, and stores

Example:
```svelte
<!-- Parent -->
<ArticleCard article={item} on:click={() => openModal(item)} />

<!-- Child -->
<script>
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
</script>
<div on:click={() => dispatch('click')}>...</div>
```

### Masonry Grid
- **React**: react-masonry-css library
- **Svelte**: CSS Grid with custom logic or use svelte-masonry library

Option 1 - Pure CSS:
```css
.masonry-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 1rem;
  grid-auto-flow: dense;
}
```

Option 2 - Install `svelte-masonry`:
```bash
npm install svelte-masonry
```

## API Compatibility

The SvelteKit frontend will use the **exact same API** as the React frontend:
- `GET /api/articles` - List articles with pagination
- `GET /api/feeds` - List all feeds
- `POST /api/feeds` - Create new feed
- `DELETE /api/feeds/:id` - Delete feed
- `POST /api/feeds/:id/refresh` - Refresh feed
- `PUT /api/articles/:id/read` - Mark as read/unread
- `PUT /api/articles/:id/star` - Star/unstar article
- `GET /api/discover?url=` - Discover feeds from website
- `GET /api/youtube/resolve?input=` - Resolve YouTube channel

No backend changes required!

## Build Configuration

### Development
```bash
cd frontend-svelte
npm run dev
```
- Runs on http://localhost:5173
- Proxies API requests to http://localhost:8080

### Production Build
```bash
cd frontend-svelte
npm run build
```
- Outputs to `build/` directory
- Static files ready to serve
- Update backend to serve from `../frontend-svelte/build` instead of `../frontend/dist`

### Backend Update
In `backend/main.go`, add a check for SvelteKit build:
```go
// Try SvelteKit build first, fall back to React
frontendPath := "../frontend-svelte/build"
if _, err := os.Stat(frontendPath); err != nil {
    frontendPath = "../frontend/dist"
}
```

## Dependencies to Install

Already installed:
- @sveltejs/kit
- @sveltejs/adapter-static
- svelte
- vite
- tailwindcss@^3
- postcss
- autoprefixer

May need to add:
```bash
npm install date-fns  # For date formatting
npm install svelte-masonry  # Optional, for masonry grid
```

## Benefits of SvelteKit Version

1. **Smaller bundle size** - Svelte compiles to vanilla JS, no runtime
2. **Better performance** - No virtual DOM diffing
3. **Simpler state management** - Built-in stores instead of external libraries
4. **Less boilerplate** - More concise component syntax
5. **Reactive by default** - `$:` reactive statements
6. **Native animations** - Built-in transitions and animations

## Timeline Estimate

- **Phase 1** (Setup): 30 minutes
- **Phase 2** (State): 30 minutes
- **Phase 3** (Components): 2-3 hours
- **Phase 4** (Main page): 30 minutes
- **Phase 5** (Styling): 1 hour
- **Phase 6** (Testing): 1 hour

**Total: 5-6 hours** for a complete, feature-equivalent SvelteKit frontend

## Next Steps

1. Complete the component implementations
2. Test thoroughly against the Go backend
3. Compare bundle sizes and performance
4. Document any Svelte-specific best practices discovered
5. Consider making SvelteKit version the default if performance gains are significant

## Notes

- Keep the React version (`frontend/`) intact as a reference
- Both frontends can coexist and use the same backend
- SvelteKit version is in `frontend-svelte/`
- No database or API changes needed
- Focus on feature parity, not new features
