# Hướng dẫn mở rộng hệ thống — Thêm một service mới (bằng tiếng Việt)

Tài liệu này mô tả các bước cụ thể để thêm một service mới vào template Go MVC hiện tại. Ví dụ: thêm service `posts` (hoặc `product`, `order`...) từ lúc tạo bảng DB, viết repository, service, controller, khai route, đến viết test và cấu hình docker.

Mục tiêu: bạn có thể làm theo từng bước sau để thêm service mới mà không phá cấu trúc hiện có.

## Tổng quan các bước

1. Thiết kế bảng DB và migration SQL.
2. Cập nhật docker-compose / .env (nếu cần) để có DB và chạy migration.
3. Thêm model ở `internal/models`.
4. Thêm repository (interface + Postgres implementation hoặc InMemory) ở `internal/repositories`.
5. Thêm service nơi chứa business logic ở `internal/services`.
6. Thêm controller (HTTP handlers) ở `internal/controllers`.
7. Đăng ký routes trong `internal/server/router.go` (hoặc một file router riêng).
8. Viết unit tests (service/repo/controller) và integration tests nếu cần.
9. Chạy migrations và khởi động dịch vụ (docker compose hoặc local).

---

## 1) Thiết kế DB và tạo migration

- Ví dụ `posts` table minimal:

```sql
CREATE TABLE posts (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author_id TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
```

- Lời khuyên: dùng công cụ migration (migrate, golang-migrate, goose, or sql-migrate). Đặt script migration vào thư mục `migrations/`.

Ví dụ dùng `migrate`:

```cmd
# cài migrate (local)
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# tạo file migration SQL
migrate create -ext sql -dir migrations create_posts_table
```

Sau đó thêm SQL vào file migration sinh ra.

## 2) Cấu hình Docker / Compose để chạy migration (tuỳ chọn)

- `docker-compose.yml` hiện có service `db` và `app`. Bạn có thể thêm một dịch vụ `migrate` tạm thời để chạy migration khi deploy hoặc tích hợp migration vào bước khởi động app (thực hiện cẩn thận!).

Ví dụ một service `migrate` trong `docker-compose.yml` (tham khảo):

```yaml
  migrate:
    image: migrate/migrate
    command: -path=/migrations -database "${DB_URL}" up
    volumes:
      - ./migrations:/migrations
    depends_on:
      - db
    env_file: .env
```

Trong `.env` bạn cần cung cấp `DB_URL` (ví dụ `postgres://postgres:postgres@db:5432/template?sslmode=disable`). Hoặc dùng `internal/util/config.LoadDB().PostgresDSN()` để xây DSN trong code.

## 3) Thêm model

- Tạo `internal/models/post.go`:

```go
package models

type Post struct {
    ID string `json:"id"`
    Title string `json:"title"`
    Body string `json:"body"`
    AuthorID string `json:"author_id"`
}
```

## 4) Thêm repository

- Tạo interface repository để dễ mock trong test:

`internal/repositories/post_repo.go`

```go
package repositories

import "template-go-jwt/internal/models"

type PostRepository interface {
    Create(p *models.Post) error
    GetByID(id string) (*models.Post, error)
    ListByAuthor(authorID string) ([]*models.Post, error)
}
```

- Implement in-memory for dev/tests: `internal/repositories/post_repo_mem.go`.
- Implement Postgres-backed repo later (using `database/sql` + `pgx` or `sqlx`). Use prepared statements and context.

Example Postgres repo skeleton (pseudo):

```go
type PostRepoPG struct { db *sql.DB }

func (r *PostRepoPG) Create(p *models.Post) error {
  _, err := r.db.ExecContext(ctx, "INSERT INTO posts (id,title,body,author_id) VALUES ($1,$2,$3,$4)", p.ID, p.Title, p.Body, p.AuthorID)
  return err
}
```

## 5) Thêm service

- Tạo `internal/services/post_service.go` để chứa logic như validate, business rules, event emit, etc.

```go
package services

import (
  "template-go-jwt/internal/models"
  "template-go-jwt/internal/repositories"
)

type PostService struct { repo repositories.PostRepository }

func NewPostService(r repositories.PostRepository) *PostService { return &PostService{repo: r} }

func (s *PostService) Create(p *models.Post) error {
  // validate -> repo.Create
}
```

Viết unit test cho service (happy path + validation failures).

## 6) Thêm controller (HTTP handlers)

- Tạo `internal/controllers/post_controller.go`:

```go
package controllers

import (
  "encoding/json"
  "net/http"

  "template-go-jwt/internal/response"
  "template-go-jwt/internal/services"
)

type PostController struct { svc *services.PostService }

func NewPostController(s *services.PostService) *PostController { return &PostController{svc:s} }

func (c *PostController) Create(w http.ResponseWriter, r *http.Request) {
  var req models.Post
  if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    response.Error(w, http.StatusBadRequest, "invalid body")
    return
  }
  if err := c.svc.Create(&req); err != nil { response.Error(w, http.StatusInternalServerError, err.Error()); return }
  response.JSON(w, http.StatusCreated, req)
}
```

## 7) Đăng ký routes

- Mở `internal/server/router.go` và đăng ký route mới. Ví dụ:

```go
mux.HandleFunc("/posts", postCtrl.Create) // POST /posts
mux.HandleFunc("/posts/", postCtrl.GetByID) // GET /posts/{id}
```

- Nếu route cần auth: wrap bằng `middleware.AuthMiddleware(cfg.JWTSecret, handler)` hoặc `RequireRole(...)` nếu cần role-based.

## 8) Viết tests

- Unit tests:
  - Repository in-memory tests: Create/Get/List
  - Service tests: validation, error paths
  - Controller tests: handler tests using httptest.Server / httptest.NewRecorder

Ví dụ test handler:

```go
func TestCreatePostHandler(t *testing.T) {
  repo := repositories.NewInMemoryPostRepo()
  svc := services.NewPostService(repo)
  ctrl := controllers.NewPostController(svc)

  body := `{"id":"p1","title":"hi","body":"b","author_id":"alice"}`
  req := httptest.NewRequest("POST", "/posts", strings.NewReader(body))
  w := httptest.NewRecorder()
  ctrl.Create(w, req)
  if w.Code != http.StatusCreated { t.Fatalf("expected 201 got %d", w.Code) }
}
```

## 9) Migrations và khởi chạy

- Khi migration đã sẵn sàng, bạn có thể:
  - Chạy migration local bằng `migrate` hoặc script SQL trực tiếp.
  - Thêm step migration vào CI/CD hoặc docker-compose `migrate` service (như phần 2).

## 10) Docker / CI notes

- Nếu bạn muốn app connect DB ngay khi khởi động trong compose, hãy thêm retry/wait-for logic để đợi Postgres sẵn sàng (khoảng 3-5 lần thử với sleep).
- Tốt hơn là chạy migration như bước riêng trong deploy pipeline rồi mới khởi động app.

## Checklist tóm tắt (copy/paste)

1. [ ] Tạo migration SQL và đảm bảo nó chạy thành công.
2. [ ] Tạo `internal/models/<resource>.go`.
3. [ ] Tạo `internal/repositories/<resource>_repo.go` (interface) và in-memory hoặc PG impl.
4. [ ] Tạo `internal/services/<resource>_service.go`.
5. [ ] Tạo `internal/controllers/<resource>_controller.go`.
6. [ ] Ghi route vào `internal/server/router.go`.
7. [ ] Viết tests cho repo/service/controller.
8. [ ] Chạy `go test ./...` và chạy migration.
9. [ ] Cập nhật README / docs / API spec.

## Ví dụ nhanh (curl)

1. Tạo post (giả sử không cần auth):

```bash
curl -X POST http://localhost:8080/posts -H 'Content-Type: application/json' -d '{"id":"p1","title":"Hello","body":"World","author_id":"alice"}'
```

2. Lấy post:

```bash
curl http://localhost:8080/posts/p1
```

---

Nếu bạn muốn, mình có thể: 

- tạo một mẫu `posts` feature hoàn chỉnh cho dự án (migration SQL + in-memory repo + service + controller + route + tests), hoặc
- thêm integration test mẫu dùng docker-compose để khởi Postgres và chạy migration rồi test endpoints.

Hãy cho biết bạn muốn mẫu mã hoàn chỉnh (mình sẽ tạo các file mã), hay chỉ cần tài liệu như trên. 
