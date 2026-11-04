# Giai đoạn 0 - CRUD nối thẳng Postgres

Làm được `POST /accounts` và `GET /accounts/{id}` bằng một file `main.go`, `net/http` và `pgx`, Postgres chạy bằng `docker compose`.

## Chuyện đáng kể: Postgres tắt mà API vẫn trả 404

**Vấn đề.** Khi Postgres tắt, `GET /accounts/1` vẫn trả `404 Not Found` kèm lỗi kết nối. Client sẽ hiểu là tài khoản số 1 không tồn tại, trong khi thật ra server không nói chuyện được với DB. Trước đó phiên bản đầu còn tệ theo hướng ngược lại: id không tồn tại (`/accounts/999`) lại trả `500`.

**Cách làm hỏng.** `docker compose stop` để tắt Postgres, rồi `curl -i localhost:8080/accounts/1`.

**Nguyên nhân.** Lệnh `pool.QueryRow(...).Scan(...)` trả `err != nil` ở cả hai tình huống: không có hàng nào khớp, và không nối được DB. Nhánh `if err != nil` gộp hai chuyện khác bản chất làm một nên chỉ trả được một mã trạng thái cho cả hai.

**Cách sửa.** Tách hai loại lỗi bằng `errors.Is(err, pgx.ErrNoRows)`:

- Là `pgx.ErrNoRows` (truy vấn chạy tốt nhưng không có hàng) thì client sai id, trả `404`.
- Mọi lỗi còn lại (kết nối hỏng, SQL sai, ...) là lỗi phía server: ghi lỗi thật vào log và trả `500` với câu chung chung, không đưa nội dung lỗi (user, địa chỉ DB) ra cho client.

**Đánh đổi.** Handler vẫn chỉ trả text, chưa có JSON chuẩn, và `pool` là biến cấp package để handler dùng được. Chấp nhận vì mục tiêu của giai đoạn này là tự viết được CRUD, còn việc tách tầng (repository, interface) để khi nào thật sự khó đọc mới làm.

## Những lỗi nhỏ khác đã gặp

- `INSERT` thành công nhưng `Scan` lỗi thì client nhận `500` mà hàng đã nằm trong DB. Thử lại sẽ tạo bản ghi trùng. Đây là lý do cần idempotency key ở giai đoạn 2.
- `not null` chỉ chặn `NULL`, không chặn chuỗi rỗng. JSON `{}` hay `{"owner":""}` vẫn tạo được tài khoản, nên phải kiểm tra `owner` trong handler.
- `pool, err := ...` trong `main` che mất biến `pool` cấp package, handler thấy `nil`. Phải dùng `=`.
- `Scan` và `Decode` cần địa chỉ (`&biến`) để ghi vào biến của mình.
- `QueryRow` chỉ lấy hàng đầu tiên: quên `where id = $1` thì mọi id đều trả về cùng một tài khoản mà không có lỗi nào.
