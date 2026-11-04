# momo-mini

Backend ví/thanh toán nhỏ viết bằng Go. Mục tiêu: học distributed systems bằng cách
gặp vấn đề thật, làm hỏng nó, rồi sửa. Mỗi giai đoạn là một câu chuyện kể được khi phỏng vấn.

Chạy được bằng một lệnh: `docker compose up`.

## Giai đoạn 0 - Tự viết được CRUD (nền tảng)

Lý do: 8-9 tháng học bằng cách đọc code AI viết (kiểu "store map" trong RAM) mà vẫn chưa
tự viết được CRUD. Giai đoạn này sửa cách học trước khi học thứ khó hơn.

- Một file `main.go`. Không chia package, không interface, không "store", không framework.
- `net/http` (Go >= 1.22 có route theo method + path). Kiểm tra `go version` trước.
- Nối thẳng Postgres (docker), dùng `database/sql` hoặc `pgx`.
- 2 endpoint: `POST /accounts`, `GET /accounts/{id}`. Khoảng 80-100 dòng.
- Thử bằng `curl`, rồi cố tình làm hỏng: JSON sai, id không tồn tại, tắt Postgres.
- Chỉ tách file/package khi `main.go` thật sự khó đọc.

Bước đầu tiên: `GET /ping` + Postgres chạy bằng docker.

## Giai đoạn 1 - Chuyển tiền và race condition

- `POST /transfer` ngây thơ (đọc số dư, kiểm tra, ghi).
- Test 100 goroutine song song -> thấy nó sai.
- Sửa: transaction, `SELECT ... FOR UPDATE`, lock theo thứ tự id để tránh deadlock.

## Giai đoạn 2 - Idempotency key

- Test gửi cùng một request 20 lần song song -> số dư chỉ đổi một lần.
- Cốt lõi: A không biết request timeout đã được xử lý chưa -> gửi kèm mã duy nhất, B lưu mã
  cùng transaction với việc trừ tiền (UNIQUE constraint).

## Giai đoạn 3 - Redis

- Rate limit, cache số dư.
- Tái hiện cache stampede, sửa bằng `singleflight`.
- Tự thử vì sao Redis lock đơn giản là chưa đủ.

## Giai đoạn 4 - Outbox + queue

- Ghi event vào bảng outbox cùng transaction, worker đẩy sang RabbitMQ (Kafka làm sau).
- Consumer phải chịu được nhận message 2 lần (at-least-once).
- Test bằng `kill -9` worker giữa chừng.
- Viết rõ RabbitMQ (queue, ack xong thì xóa) khác Kafka (log, mỗi group giữ offset) ở đâu.

## Giai đoạn 5 - Đo đạc

- `slog`, Prometheus, `pprof`, load test bằng k6.
- Tìm điểm nghẽn thật, sửa, ghi số trước/sau vào README.

## Giai đoạn 6 - k8s (cuối cùng)

- `kind` chạy local: readiness probe, graceful shutdown, rolling update không rớt request.

## Quy tắc cho cả repo

- Mỗi giai đoạn có test tái hiện lỗi, chạy fail trước, rồi mới sửa cho pass. Commit riêng.
- Mỗi giai đoạn có một đoạn ngắn trong `docs/`: vấn đề, cách làm hỏng, nguyên nhân, cách sửa,
  đánh đổi (vd: "tại sao ở đây chưa cần Kafka").
- Commit nhỏ và đều, không đẩy cả đống lên cuối.
- Tự gõ phần lõi và giải thích được từng dòng.
- Ước chừng mỗi giai đoạn 1-2 tuần (vừa đi làm vừa học).
