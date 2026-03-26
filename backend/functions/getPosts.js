import { promises as fs } from "fs";
import path from "path";
import { fileURLToPath } from "url";
import yaml from "js-yaml";

// 현재 파일의 디렉토리 경로를 가져옴
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// posts 디렉토리의 절대 경로
const POSTS_DIR = path.join(__dirname, "..", "posts");

/**
 * post type for list view (excludeContent = true)
 * @typedef {Object} PostListItem
 * @property {string} folder - post folder name
 * @property {string} title - post title
 * @property {string} date - post date
 * @property {Array<string>} tag - post tags
 * @property {string} summary - post summary
 * @property {string} coverImg - post cover image
 */

/**
 * post type for detail view (excludeContent = false)
 * @typedef {Object} PostDetail
 * @property {string} folder - post folder name
 * @property {string} title - post title
 * @property {string} date - post date
 * @property {Array<string>} tag - post tags
 * @property {string} content - post content
 */

/**
 * get posts from post directory
 * @param {string} targetFolder - target folder name
 * @param {boolean} excludeContent - exclude content from posts (default: false)
 * @returns {Promise<Array<PostListItem|PostDetail>>} posts from post directory (if targetFolder is not null, return only the target folder)
 */
async function getPosts(targetFolder, excludeContent = false) {
	try {
		// posts 디렉토리에 있는 폴더(포스트)들을 읽어옴

		const folders = targetFolder ? [targetFolder] : await fs.readdir(POSTS_DIR);

		if (folders.length === 0) {
			return [];
		}

		// posts 디렉토리에 있는 포스트들을 불러옴
		const posts = await Promise.all(
			folders.map(async (folder) => {
				const indexPath = path.join(POSTS_DIR, folder, "index.md");
				const content = await fs.readFile(indexPath, "utf-8");

				const yamlContent = content.split("---")[1];
				const frontMatter = yaml.load(yamlContent);

				const postData = {
					folder: folder,
					title: frontMatter.title,
					date: frontMatter.date,
					tag: frontMatter.tag || [],
				};

				if (excludeContent) {
					// 포스트 목록용: summary와 coverImg 포함, content 제외
					postData.summary = frontMatter.summary || "";
					postData.coverImg = frontMatter.coverImg || null;
				} else {
					// 개별 포스트용: content 포함, summary와 coverImg 제외
					postData.content = content.split("---").slice(2).join("---").trim();
				}

				return postData;
			})
		);

		return posts;
	} catch (error) {
		console.error("Error reading posts:", error);
		throw new Error("Failed to read posts");
	}
}

export default getPosts;
