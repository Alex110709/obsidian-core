# Synology 배포 가이드

## 🎯 올바른 설정

### docker-compose.yml

```yaml
services:
  obsidian-node:
    image: yuchanshin/obsidian-node:v1.0.0
    container_name: obsidian-node
    environment:
      - DATA_DIR=/root/data
      - TOR_ENABLED=true
    ports:
      - "18333:8333"  # P2P port
      - "18545:8545"  # RPC port
      - "9050:9050"   # Tor SOCKS proxy
    volumes:
      - /volume1/HDD_DATA/obsidian:/root/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD-SHELL", "pgrep obsidiand || exit 1"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 90s
```

## 📋 주요 변경사항

### ❌ 잘못된 볼륨 마운트
```yaml
volumes:
  - /volume1/HDD_DATA/obsidian:/root  # ❌ 이렇게 하면 바이너리 덮어씌움
```

### ✅ 올바른 볼륨 마운트
```yaml
environment:
  - DATA_DIR=/root/data  # 데이터 디렉토리 지정
volumes:
  - /volume1/HDD_DATA/obsidian:/root/data  # 데이터만 마운트
```

## 🚀 배포 단계

### 1. SSH로 Synology 접속

```bash
ssh admin@your-synology-ip
```

### 2. 데이터 디렉토리 생성

```bash
sudo mkdir -p /volume1/HDD_DATA/obsidian
sudo chmod 755 /volume1/HDD_DATA/obsidian
```

### 3. Container Manager에서 배포

1. **Project** 탭 열기
2. **Create** 클릭
3. 프로젝트 이름: `obsidian-node`
4. 위의 `docker-compose.yml` 내용 붙여넣기
5. **Build** 클릭

### 4. 로그 확인

```bash
sudo docker logs -f obsidian-node
```

예상 출력:
```
Starting Obsidian Node...
Network: mainnet
Block Size Limit: 6000000 bytes
Started Tor process (PID: 123)
Tor process started successfully
Tor enabled via proxy: 127.0.0.1:9050
RPC server listening on 0.0.0.0:8545
Miner started. Mining on CPU...
```

## 🔧 환경변수

| 변수 | 설명 | 기본값 |
|------|------|--------|
| `DATA_DIR` | 데이터 저장 경로 | `.` (현재 디렉토리) |
| `TOR_ENABLED` | Tor 활성화 | `false` |
| `RPC_ADDR` | RPC 서버 주소 | `0.0.0.0:8545` |

## 📂 데이터 구조

```
/volume1/HDD_DATA/obsidian/
├── obsidian.db          # 블록체인 데이터베이스
└── tor/                 # Tor 데이터 (TOR_ENABLED=true 시)
    ├── torrc            # Tor 설정
    └── ...
```

## 🧪 테스트

### RPC API 호출
```bash
curl -X POST http://your-synology-ip:18545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockchaininfo","params":[],"id":1}'
```

### 블록 높이 확인
```bash
curl -X POST http://your-synology-ip:18545 \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc":"2.0","method":"getblockcount","params":[],"id":1}'
```

## 🐛 문제 해결

### "no such file or directory" 에러
- 이전 이미지 캐시 문제
- 해결: `sudo docker rmi yuchanshin/obsidian-node:v1.0.0 && sudo docker pull yuchanshin/obsidian-node:v1.0.0`

### Tor 시작 실패
- 권한 문제일 수 있음
- 해결: `user: "0:0"` 추가 또는 제거

### 데이터베이스 권한 에러
- 디렉토리 권한 문제
- 해결: `sudo chmod -R 755 /volume1/HDD_DATA/obsidian`

## 📦 버전 정보

- **현재 버전**: v1.0.0
- **이미지**: `yuchanshin/obsidian-node:v1.0.0`
- **플랫폼**: linux/amd64, linux/arm64
