# Hướng dẫn mở rộng hệ thống — Thêm một service mới (bằng tiếng Việt)

Tài liệu này cập nhật hướng dẫn mở rộng dự án theo trạng thái hiện tại: hệ thống dùng GORM cho ORM, và file-based migrations (thư mục `migrations/`) được chạy bằng `golang-migrate` thông qua binary `cmd/migrate` hoặc service `migrate` trong `docker-compose.yml`.

Mục tiêu: bạn có thể thêm một resource/service (ví dụ `posts`) từ thiết kế DB, tạo file migrations, đến viết model/repo/service/controller, đăng ký route và test.

## Tổng quan các bước

1. Thiết kế bảng DB và tạo file-based migrations (thư mục `migrations/`).
2. (Tuỳ chọn) Cập nhật `docker-compose.yml` / `.env` để chạy DB và job `migrate`.
3. Thêm model ở `internal/models` với tag GORM.
4. Thêm repository (interface + in-memory + GORM Postgres impl) ở `internal/repositories`.
5. Thêm service (business logic) ở `internal/services`.
6. Thêm controller (HTTP handlers) ở `internal/controllers`.
7. Đăng ký routes trong `internal/server/router.go`.
8. Viết unit tests và (tuỳ) integration tests.

---

## 1) File-based migrations (golang-migrate)

Project hiện có thư mục `migrations/`. Migrations là cặp file `.up.sql` / `.down.sql` (ví dụ `1_create_users.up.sql` và `1_create_users.down.sql`).

Tạo migration mới (cách nhanh, dùng migrate CLI):

1. Cài CLI (local dev):

```cmd
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

2. Tạo migration (ví dụ: tạo posts):

```cmd
migrate create -ext sql -dir migrations create_posts_table
```

Lệnh trên sẽ tạo hai file có tiền tố số phiên bản trong `migrations/`: `000002_create_posts_table.up.sql` và `000002_create_posts_table.down.sql` (số có thể khác). Chỉnh sửa `.up.sql` để thêm DDL, và `.down.sql` để rollback.

Ví dụ up:

```sql
CREATE TABLE IF NOT EXISTS posts (
  id TEXT PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author_id TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
```

Ví dụ down:

```sql
DROP TABLE IF EXISTS posts;
```

3. Chạy migration:

- Từ code (đã thêm `cmd/migrate`):

```cmd
go run ./cmd/migrate
```

`cmd/migrate` dùng `golang-migrate` để chạy các file trong `migrations/` và áp dụng chúng vào DB URL lấy từ `.env` (hàm `internal/util/config.LoadDB().PostgresURL()`).

- Hoặc dùng docker compose (khuyến nghị cho môi trường đúng):

```cmd
docker compose build app
docker compose up -d db
docker compose run --rm migrate
```

`migrate` service trong `docker-compose.yml` đã cấu hình để chạy `/app/migrate` (binary trong image) và dùng `migrations/` đã được COPY vào image builder context. Nếu bạn chỉ cần rollback, bạn có `docker compose run --rm migrate down` (hiện `cmd/migrate` hỗ trợ tham số `down`).

4. Lưu ý khi xây DB URL: nếu mật khẩu chứa ký tự đặc biệt, hãy escape nó (hoặc dùng biến môi trường an toàn). File `internal/util/config` cung cấp `PostgresURL()` helper.

## 2) Thêm model

Tạo `internal/models/post.go` với tag GORM, ví dụ:

```go
package models

type Post struct {
  ID       string `gorm:"primaryKey" json:"id"`
  Title    string `json:"title"`
  Body     string `json:"body"`
  AuthorID string `json:"author_id"`
}
```

Bạn có thể dùng GORM's `AutoMigrate` during development, nhưng lưu ý: schema changes should be driven by SQL migrations for production.

## 3) Repository

- Tạo interface `PostRepository` trong `internal/repositories/post_repo.go`.
- Implement an in-memory repo for tests and a GORM-based repo for production.

GORM repo skeleton:

```go
type PostRepoGorm struct { db *gorm.DB }

func NewPostRepoGorm(db *gorm.DB) *PostRepoGorm { return &PostRepoGorm{db: db} }

func (r *PostRepoGorm) Create(p *models.Post) error {
  return r.db.Create(p).Error
}

func (r *PostRepoGorm) GetByID(id string) (*models.Post, error) {
  var p models.Post
  if err := r.db.First(&p, "id = ?", id).Error; err != nil {
    return nil, err
  }
  return &p, nil
}
```

## 4) Service & Controller

- Service: `internal/services/post_service.go` chứa business logic and calls repo.
- Controller: `internal/controllers/post_controller.go` contains HTTP handlers and uses `response.JSON` / `response.Error` helpers.

## 5) Register routes

Add routes in `internal/server/router.go`:

```go
mux.HandleFunc("/posts", postCtrl.Create) // POST
mux.HandleFunc("/posts/", postCtrl.GetByID) // GET /posts/{id}
```

Wrap handlers with `middleware.AuthMiddleware` or `middleware.RequireRole` if route requires authentication/authorization.

## 6) Tests

- Unit tests for repo/service/controller (use in-memory repo for controller/service tests).
- Integration tests: use docker-compose to start a real Postgres, run `migrate`, then run HTTP tests against the running app.

## 7) Migration creation quick reference

Local dev with migrate CLI:

```cmd
go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
migrate create -ext sql -dir migrations create_posts_table
```

Then edit the generated `.up.sql` and `.down.sql`, then run:

```cmd
go run ./cmd/migrate
```

Or via Docker (build once):

```cmd
docker compose build app
docker compose up -d db
docker compose run --rm migrate
```

## Checklist (copy/paste)

1. [ ] `migrate create -ext sql -dir migrations <name>`
2. [ ] Fill `.up.sql` and `.down.sql` with DDL
3. [ ] `go run ./cmd/migrate` (or `docker compose run --rm migrate`)
4. [ ] Add model struct in `internal/models`
5. [ ] Add repository interface + GORM implementation in `internal/repositories`
6. [ ] Add service in `internal/services`
7. [ ] Add controller in `internal/controllers` and register route
8. [ ] Add tests and run `go test ./...`

---

Nếu bạn muốn, tôi có thể:

- Tạo migration mẫu `posts` (file up/down SQL) và tạo `internal/models/post.go`, `internal/repositories/post_repo_gorm.go`, `internal/services/post_service.go`, `internal/controllers/post_controller.go` và route registration + tests. Tôi sẽ thực hiện từng phần và chạy `go test ./...` sau mỗi bước.
- Hoặc chỉ tạo migration mẫu và hướng dẫn bạn chạy nó.

Chọn 1 trong 2 (tạo mẫu đầy đủ cho `posts` hoặc chỉ tạo migration mẫu) để tôi làm tiếp.
