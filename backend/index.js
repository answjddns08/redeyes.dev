import express, { json } from "express";
import cors from "cors";
import compression from "compression";
import postRouter from "./router/postRoute.js";
import tagRouter from "./router/tagRoute.js";

const app = express();

// Gzip 압축 활성화 (응답 크기 줄이기)
app.use(
	compression({
		filter: (req, res) => {
			if (req.headers["x-no-compression"]) {
				return false;
			}
			return compression.filter(req, res);
		},
		level: 6, // 압축 레벨 (1-9, 6이 기본값)
		threshold: 1024, // 1KB 이상일 때만 압축
	})
);

// CORS 설정
app.use(
	cors({
		origin: [
			"https://blog.redeyes.dev",
			"http://localhost:5173",
			"https://redeyes.dev",
		],
		methods: ["GET", "POST", "DELETE", "PUT"],
		allowedHeaders: ["Content-Type", "Authorization"],
		credentials: true,
	})
);

// 성능 최적화 미들웨어
app.use((req, res, next) => {
	// 정적 파일(이미지)에 대한 캐시 헤더 설정
	if (req.path.includes("/images/")) {
		res.set("Cache-Control", "public, max-age=604800"); // 일주일 캐시
		res.set("ETag", `"${Date.now()}"`);
	}

	next();
});

// JSON 파싱 미들웨어
app.use(json({ limit: "10mb" })); // 파일 크기 제한

/*
  JSON 파싱 미들웨어:
  json 형식의 요청을 받아 그걸 javascript 객체로 변환해줌
*/

// 요청 로깅 (개발용)
if (process.env.NODE_ENV !== "production") {
	app.use((req, res, next) => {
		console.log(`${new Date().toISOString()} - ${req.method} ${req.path}`);
		next();
	});
}

app.use("/api/tags", tagRouter);
app.use("/api/posts", postRouter);

// 에러 핸들링 미들웨어
app.use((err, req, res, next) => {
	console.error("Error:", err);
	res.status(500).json({
		message: "Internal Server Error",
		...(process.env.NODE_ENV !== "production" && { error: err.message }),
	});
});

// 404 핸들링
app.use((req, res) => {
	res.status(404).json({ message: "Not Found" });
});

const PORT = 5000;
const HOST = "127.0.0.1";

app.listen(PORT, HOST, () => {
	console.log(`Server is running on ${HOST}:${PORT}`);
});
