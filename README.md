# webhook-go

간단한 Go 기반 webhook 서버입니다.

## 로컬 실행

```bash
go run .
```

기본 포트는 `8080`이며, `PORT` 환경 변수로 변경할 수 있습니다.

## 엔드포인트

- `GET /healthz` : 상태 확인
- `POST /webhook` : webhook payload 수신

## Docker 이미지 빌드

```bash
docker build -t webhook-go:latest .
```

## Docker 컨테이너 실행

```bash
docker run --rm -p 8080:8080 webhook-go:latest
```
