# Kellog Blog 📝

Vue.js 3와 Express.js로 구축한 개인 블로그 프로젝트입니다. 마크다운 기반의 포스트 작성과 태그 시스템, 다크 모드를 지원하는 간단한 블로그입니다

## 🚀 주요 기능

- **마크다운 기반 포스트**: `.md` 파일로 블로그 포스트 작성
- **태그 시스템**: 포스트 카테고리화 및 필터링
- **검색 기능**: 포스트 제목, 내용 기반 검색
- **다크 모드**: 시스템 설정 감지 및 수동 토글
- **목차 기능**: 포스트의 헤딩 기반 네비게이션
- **캐싱 시스템**: localStorage를 활용한 성능 최적화

## 🛠 기술 스택

### Frontend
- **Vue.js 3** - Composition API
- **Vite** - 빌드 도구
- **Vue Router** - SPA 라우팅
- **Pinia** - 상태 관리
- **Tailwind CSS** - 스타일링
- **Font Awesome** - 아이콘
- **Axios** - HTTP 클라이언트

### Backend
- **Node.js** - 런타임
- **Express.js** - 웹 프레임워크
- **Markdown-it** - 마크다운 파싱
- **CORS** - Cross-Origin 지원

## 📁 프로젝트 구조

```
├── frontend/          # Vue.js 클라이언트
│   ├── src/
│   │   ├── components/    # 재사용 가능한 컴포넌트
│   │   ├── stores/        # Pinia 상태 관리
│   │   ├── views/         # 페이지 컴포넌트
│   │   └── router/        # 라우팅 설정
│   └── public/            # 정적 파일
├── backend/           # Express.js 서버
│   ├── controller/        # API 컨트롤러
│   ├── functions/         # 유틸리티 함수
│   ├── posts/            # 마크다운 포스트 파일
│   └── router/           # API 라우트
└── README.md
```

## 🏃‍♂️ 시작하기

### 필수 요구사항
- Node.js 16.0+ 
- npm 또는 yarn

### 설치 및 실행

1. **저장소 클론**
```bash
git clone https://github.com/answjddns08/blogProject.git
```

2. **의존성 설치**
```bash
# Frontend 의존성 설치
cd frontend
npm install

# Backend 의존성 설치
cd ../backend
npm install
```

3. **개발 서버 실행**
```bash
# Backend 서버 실행
cd backend
npm start

# Frontend 개발 서버 실행
cd frontend
npm run dev
```

4. **브라우저에서 확인**
- Frontend: http://localhost:5173
- Backend API: http://localhost:3000

## 📝 포스트 작성하기

1. `backend/posts/` 디렉토리에 새 폴더 생성
2. 폴더 내에 `index.md` 파일 생성
3. 마크다운 형식으로 포스트 작성:

```markdown
---
title: "포스트 제목"
summary: |
    포스트 요약
tag:
    - Tag1
    - Tag2
date: 1972/7/11
coverImg: "cover.png"  # 선택사항
---

# 포스트 내용

마크다운으로 작성된 포스트 내용...
```

## 🎨 커스터마이징

### 테마 색상 변경
`frontend/src/assets/style.css`에서 CSS 변수 수정:

```css
:root {
  --primary-color: #your-color;
  --accent-color: #your-accent;
  /* ... */
}
```

### 컴포넌트 수정
- 네비게이션: `frontend/src/components/NavBar.vue`
- 포스트 목록: `frontend/src/components/ShowPosts.vue`
- 태그 시스템: `frontend/src/components/showTags.vue`


## 🚀 배포

### Frontend 빌드
```bash
cd frontend
npm run build
```

### Production 환경
- Frontend: `frontend/dist/` 폴더를 정적 호스팅
- Backend: Node.js 서버 환경에 배포
