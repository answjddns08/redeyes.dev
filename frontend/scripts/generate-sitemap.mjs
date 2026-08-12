/**
 * sitemap.xml 자동 생성 스크립트
 * - 정적 페이지(/, /about)와 API에서 가져온 글 목록을 기반으로 public/sitemap.xml 생성
 * - 빌드 전에 실행됨 (npm run build 시 prebuild로, npm run deploy 시 직접 호출)
 */
import { writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const BASE_URL = "https://blog.redeyes.dev";
const POSTS_API = `${BASE_URL}/api/posts`;

const escapeXml = (str) =>
	str
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&apos;");

const urls = [
	{ path: "/", changefreq: "daily", priority: "1.0" },
	{ path: "/about", changefreq: "monthly", priority: "0.5" },
];

try {
	const res = await fetch(POSTS_API, {
		headers: { "User-Agent": "sitemap-generator" },
	});
	if (!res.ok) throw new Error(`API 응답 ${res.status}`);
	const posts = await res.json();

	for (const post of posts) {
		urls.push({
			path: `/posts/${encodeURIComponent(post.folder)}`,
			lastmod: post.date || undefined,
			changefreq: "monthly",
			priority: "0.8",
		});
	}
	console.log(`sitemap: 글 ${posts.length}개 포함`);
} catch (err) {
	console.warn(`sitemap: 글 목록을 가져오지 못해 정적 페이지만 포함 (${err.message})`);
}

const urlset = urls
	.map(
		(u) => `  <url>
    <loc>${BASE_URL}${escapeXml(u.path)}</loc>${u.lastmod ? `\n    <lastmod>${u.lastmod}</lastmod>` : ""}${u.changefreq ? `\n    <changefreq>${u.changefreq}</changefreq>` : ""}${u.priority ? `\n    <priority>${u.priority}</priority>` : ""}
  </url>`,
	)
	.join("\n");

const xml = `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
${urlset}
</urlset>
`;

await writeFile(resolve("public/sitemap.xml"), xml, "utf8");
console.log(`sitemap.xml 생성 완료 (${urls.length}개 URL)`);
