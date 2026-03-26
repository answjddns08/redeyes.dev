import bringPosts from "../functions/getPosts.js";
import cache from "../utils/cache.js";

/*
    디렉토리에서 포스트들을 가져오고
    그 포스트들의 태그들만 가져옴
*/

async function getTags(req, res) {
	try {
		const cacheKey = "tags_all";

		// 캐시에서 먼저 확인
		const cachedTags = cache.get(cacheKey);
		if (cachedTags) {
			console.log("Cache hit for tags");
			res.json(cachedTags);
			return;
		}

		console.log("Cache miss for tags, fetching from disk...");

		const posts = await bringPosts();
		const tags = new Set(); // Set(집합)을 사용하여 중복된 태그 제거

		posts.forEach((post) => {
			post.tag.forEach((tag) => {
				tags.add(tag);
			});
		});

		const tagsArray = Array.from(tags);

		// 결과를 캐시에 저장
		cache.set(cacheKey, tagsArray);

		res.json(tagsArray);
	} catch (error) {
		console.error("Error getting tags:", error);
		res.status(500).json({ message: "Failed to get tags" });
	}
}

export { getTags };
