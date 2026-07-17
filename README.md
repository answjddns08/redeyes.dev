# Redeyes.dev

[redeyes.dev](https://blog.redeyes.dev) is my personal blog, built with [Vue.js](https://vuejs.org) and [Go](https://go.dev/).

## Overview

This repository contains both the frontend and backend for a self-hosted blog.
The frontend is a Vue 3 single-page app, and the backend is a Go API that stores rendered posts in SQLite.

## Features

- Search posts by keyword or tag from the top navigation bar.
- Browse posts with infinite scroll on the home page.
- Filter posts by tags and open posts by slug.
- View post detail pages with headings, next/previous navigation, and SEO-friendly structured data.
- Toggle light and dark themes, with the preference saved in local storage.
- Upload posts with images and delete posts through authenticated backend endpoints. (you can upload with making obsidian custom export plugin)
- Serve post images directly from the backend.

## Stack

### Frontend

- Vue 3
- Vue Router
- Pinia
- Tailwind CSS 4
- Axios
- Font Awesome

### Backend

- Go 1.26
- SQLite
- Goldmark for Markdown rendering

### Deployment

- Self-hosted
- Raspberry Pi 5 (4GB)
- Raspberry Pi OS Lite
- Nginx reverse proxy

## Project Structure

- `frontend/` contains the Vue app.
- `backend/` contains the Go API and SQLite store.
- `nginx/` contains the sample reverse proxy configuration.

## Frontend Behavior

The home page loads posts and tags from the backend, then caches them in `localStorage` for faster repeat visits.
The post list supports query-based search, tag filtering, and incremental loading as you scroll.

The post detail page renders Markdown content, shows post metadata, and links to neighboring posts.

## Backend API

The backend exposes the following endpoints:

- `GET /api/posts` returns the post list.
- `GET /api/posts?search=...` filters posts by keyword or tag.
- `GET /api/posts/:slug` returns a single post.
- `GET /api/tags` returns the unique tag list.
- `POST /api/upload` uploads a post and its images.
- `DELETE /api/posts/:slug` deletes a post and its stored images.
- `GET /api/posts/images/...` serves uploaded post images.

Uploads can be protected with `BLOG_ADMIN_TOKEN`. If the token is set, requests must send `Authorization: Bearer <token>`.
The token is stored in `backend/.env` and is not included in this repository.

```
BLOG_ADMIN_TOKEN=your-token-here
```

## Data Storage

Posts are stored in SQLite under `backend/cache/blog.db`.
When a post is uploaded, the backend parses the Markdown, stores the rendered HTML, and saves post metadata such as title, date, tags, summary, and cover image.

Uploaded images are stored under `backend/images/`.

## Local Development

### Prerequisites

- Node.js
- Go 1.26
- Nginx, if you want to use the provided reverse proxy setup

### Frontend

```bash
cd frontend
npm install
npm run dev
```

### Backend

```bash
cd backend
go run main.go
```

The backend listens on `:5000` by default. Set the `PORT` environment variable if you want to use a different port.

## Nginx

The sample configuration in `nginx/blog.conf` serves the built frontend and proxies `/api` requests to the Go backend.
It also enables history mode routing for Vue and increases the request body limit for post uploads.

## License

This project is licensed under the MIT License.

You are free to modify, redistribute, and use the source code commercially, including for ad-supported deployments.
Blog content and database data are not included in this repository.
See [LICENSE](LICENSE) for details.
