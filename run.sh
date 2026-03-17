#!/bin/bash

# ─────────────────────────────────────────
#  Standup Bot — run.sh
#  Usage: ./run.sh [dev|backend|frontend|stop]
# ─────────────────────────────────────────

BACKEND_DIR="./backend"
FRONTEND_DIR="./frontend"
BACKEND_PID_FILE=".backend.pid"
FRONTEND_PID_FILE=".frontend.pid"
BACKEND_PORT=8080
FRONTEND_PORT=3000

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# ── Dependency checks ──────────────────────────────────────────

check_dependencies() {
  local missing=0

  if ! command -v go &>/dev/null; then
    echo -e "${RED}✗ Go is not installed.${NC} Download: https://go.dev/dl/"
    missing=1
  else
    echo -e "${GREEN}✓ Go${NC} $(go version | awk '{print $3}')"
  fi

  if ! command -v node &>/dev/null; then
    echo -e "${RED}✗ Node.js is not installed.${NC} Download: https://nodejs.org/"
    missing=1
  else
    echo -e "${GREEN}✓ Node.js${NC} $(node --version)"
  fi

  if ! command -v npm &>/dev/null; then
    echo -e "${RED}✗ npm is not installed.${NC}"
    missing=1
  else
    echo -e "${GREEN}✓ npm${NC} $(npm --version)"
  fi

  if [ $missing -ne 0 ]; then
    echo ""
    echo -e "${RED}Missing dependencies. Please install them and try again.${NC}"
    exit 1
  fi

  # Check for a local AI — warn but don't block
  local ai_found=0
  for port in 8080 11434 1234 8081; do
    if curl -s --max-time 0.8 "http://localhost:$port/v1/models" &>/dev/null || \
       curl -s --max-time 0.8 "http://localhost:$port/api/tags" &>/dev/null; then
      echo -e "${GREEN}✓ Local AI detected${NC} on port $port"
      ai_found=1
      break
    fi
  done

  if [ $ai_found -eq 0 ]; then
    echo -e "${YELLOW}⚠ No local AI detected.${NC} Start Ollama, LM Studio, or mlc-llm before generating a standup."
    echo -e "  → Ollama:    https://ollama.com"
    echo -e "  → LM Studio: https://lmstudio.ai"
  fi
}

# ── Backend ────────────────────────────────────────────────────

start_backend() {
  echo -e "\n${BLUE}▶ Starting backend${NC} (port $BACKEND_PORT)..."

  if [ ! -d "$BACKEND_DIR" ]; then
    echo -e "${RED}✗ '$BACKEND_DIR' folder not found.${NC}"
    exit 1
  fi

  if [ ! -f "$BACKEND_DIR/cmd/main.go" ]; then
    echo -e "${RED}✗ backend/cmd/main.go not found.${NC}"
    exit 1
  fi

  cd "$BACKEND_DIR"

  # Install Go dependencies if needed
  if [ -f "go.mod" ]; then
    go mod tidy &>/dev/null
  fi

  go run ./cmd/main.go &
  BACKEND_PID=$!
  cd - > /dev/null

  echo $BACKEND_PID > "$BACKEND_PID_FILE"
  echo -e "${GREEN}✓ Backend started${NC} (PID: $BACKEND_PID) → http://localhost:$BACKEND_PORT"
}

# ── Frontend ───────────────────────────────────────────────────

start_frontend() {
  echo -e "\n${BLUE}▶ Starting frontend${NC} (port $FRONTEND_PORT)..."

  if [ ! -d "$FRONTEND_DIR" ]; then
    echo -e "${RED}✗ '$FRONTEND_DIR' folder not found.${NC}"
    exit 1
  fi

  cd "$FRONTEND_DIR"

  # Install npm packages if node_modules is missing
  if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}  Installing npm packages...${NC}"
    npm install --silent
  fi

  # Write .env.local if it doesn't exist
  if [ ! -f ".env.local" ]; then
    echo "NEXT_PUBLIC_API_URL=http://localhost:$BACKEND_PORT" > .env.local
    echo -e "${GREEN}  ✓ Created .env.local${NC}"
  fi

  npm run dev &
  FRONTEND_PID=$!
  cd - > /dev/null

  echo $FRONTEND_PID > "$FRONTEND_PID_FILE"
  echo -e "${GREEN}✓ Frontend started${NC} (PID: $FRONTEND_PID) → http://localhost:$FRONTEND_PORT"
}

# ── Stop ───────────────────────────────────────────────────────

stop_all() {
  echo -e "\n${BLUE}■ Stopping services...${NC}"
  local stopped=0

  if [ -f "$BACKEND_PID_FILE" ]; then
    BACKEND_PID=$(cat "$BACKEND_PID_FILE")
    if kill "$BACKEND_PID" 2>/dev/null; then
      echo -e "${GREEN}✓ Backend stopped${NC} (PID: $BACKEND_PID)"
      stopped=1
    fi
    rm -f "$BACKEND_PID_FILE"
  fi

  if [ -f "$FRONTEND_PID_FILE" ]; then
    FRONTEND_PID=$(cat "$FRONTEND_PID_FILE")
    if kill "$FRONTEND_PID" 2>/dev/null; then
      echo -e "${GREEN}✓ Frontend stopped${NC} (PID: $FRONTEND_PID)"
      stopped=1
    fi
    rm -f "$FRONTEND_PID_FILE"
  fi

  # Fallback: kill by port if PID files are stale
  if [ $stopped -eq 0 ]; then
    echo -e "${YELLOW}No PID files found. Attempting to kill by port...${NC}"
    lsof -ti:$BACKEND_PORT | xargs kill -9 2>/dev/null && echo -e "${GREEN}✓ Port $BACKEND_PORT cleared${NC}"
    lsof -ti:$FRONTEND_PORT | xargs kill -9 2>/dev/null && echo -e "${GREEN}✓ Port $FRONTEND_PORT cleared${NC}"
  fi

  echo -e "${GREEN}Done.${NC}"
}

# ── Help ───────────────────────────────────────────────────────

print_help() {
  echo ""
  echo -e "${BLUE}Standup Bot — run.sh${NC}"
  echo ""
  echo "  Usage: ./run.sh [command]"
  echo ""
  echo "  Commands:"
  echo "    dev        Start both backend and frontend (default)"
  echo "    backend    Start only the Go backend"
  echo "    frontend   Start only the Next.js frontend"
  echo "    stop       Stop all running services"
  echo "    help       Show this message"
  echo ""
  echo "  Once running:"
  echo "    Frontend  → http://localhost:$FRONTEND_PORT"
  echo "    Backend   → http://localhost:$BACKEND_PORT"
  echo ""
}

# ── Main ───────────────────────────────────────────────────────

COMMAND=${1:-dev}

case $COMMAND in
  dev)
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BLUE}  Standup Bot — Starting...${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    check_dependencies
    start_backend
    sleep 1  # Give backend a moment before frontend starts
    start_frontend
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}  All services running!${NC}"
    echo -e "${GREEN}  Open → http://localhost:$FRONTEND_PORT${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "  Press ${YELLOW}Ctrl+C${NC} or run ${YELLOW}./run.sh stop${NC} to quit."
    echo ""
    wait
    ;;
  backend)
    check_dependencies
    start_backend
    wait
    ;;
  frontend)
    check_dependencies
    start_frontend
    wait
    ;;
  stop)
    stop_all
    ;;
  help|--help|-h)
    print_help
    ;;
  *)
    echo -e "${RED}Unknown command: $1${NC}"
    print_help
    exit 1
    ;;
esac