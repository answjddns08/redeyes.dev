/**
 * 간단한 인메모리 캐시 구현
 */
class MemoryCache {
	constructor(maxSize = 100, ttl = 86400000) { // 기본 하루 TTL
		this.cache = new Map();
		this.maxSize = maxSize;
		this.ttl = ttl; // Time To Live in milliseconds
	}

	/**
	 * 캐시에서 값을 가져옴
	 * @param {string} key 
	 * @returns {any|null}
	 */
	get(key) {
		const item = this.cache.get(key);
		
		if (!item) {
			return null;
		}

		// TTL 확인
		if (Date.now() > item.expiry) {
			this.cache.delete(key);
			return null;
		}

		// 액세스 시간 업데이트 (LRU)
		item.lastAccessed = Date.now();
		return item.value;
	}

	/**
	 * 캐시에 값을 저장
	 * @param {string} key 
	 * @param {any} value 
	 */
	set(key, value) {
		// 캐시 크기 제한 확인
		if (this.cache.size >= this.maxSize) {
			this.evictLRU();
		}

		const item = {
			value,
			expiry: Date.now() + this.ttl,
			lastAccessed: Date.now()
		};

		this.cache.set(key, item);
	}

	/**
	 * LRU (Least Recently Used) 항목 제거
	 */
	evictLRU() {
		let oldestKey = null;
		let oldestTime = Date.now();

		for (const [key, item] of this.cache) {
			if (item.lastAccessed < oldestTime) {
				oldestTime = item.lastAccessed;
				oldestKey = key;
			}
		}

		if (oldestKey) {
			this.cache.delete(oldestKey);
		}
	}

	/**
	 * 캐시에서 키 삭제
	 * @param {string} key 
	 */
	delete(key) {
		return this.cache.delete(key);
	}

	/**
	 * 캐시 전체 삭제
	 */
	clear() {
		this.cache.clear();
	}

	/**
	 * 만료된 항목들 정리
	 */
	cleanup() {
		const now = Date.now();
		for (const [key, item] of this.cache) {
			if (now > item.expiry) {
				this.cache.delete(key);
			}
		}
	}

	/**
	 * 캐시 상태 정보
	 */
	getStats() {
		return {
			size: this.cache.size,
			maxSize: this.maxSize,
			keys: Array.from(this.cache.keys())
		};
	}
}

// 싱글톤 패턴으로 캐시 인스턴스 생성
const cache = new MemoryCache(20, 86400000);

// 정기적으로 만료된 캐시 정리
setInterval(() => {
	cache.cleanup();
}, 43200000); // 12시간마다 실행

export default cache;
