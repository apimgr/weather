#!/usr/bin/env bash
# Test script that runs the server with isolated temp directory

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Create temp directory for testing
TEST_DIR="${TMPDIR:-/tmp}/weather-test-$$"
mkdir -p "$TEST_DIR"

echo -e "${BLUE}🧪 Weather Service Test Server${NC}"
echo -e "${BLUE}================================${NC}"
echo -e "Test directory: ${YELLOW}$TEST_DIR${NC}"
echo ""

# Cleanup function
cleanup() {
    echo ""
    echo -e "${YELLOW}🧹 Cleaning up...${NC}"
    if [ -n "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
    if [ "$KEEP_TEMP" != "1" ]; then
        rm -rf "$TEST_DIR"
        echo -e "${GREEN}✅ Temp directory removed${NC}"
    else
        echo -e "${YELLOW}📁 Temp directory kept: $TEST_DIR${NC}"
    fi
}

trap cleanup EXIT INT TERM

# Build if needed
if [ ! -f "./weather" ]; then
    echo -e "${BLUE}🔨 Building binary...${NC}"
    go build -o weather ./src || {
        echo -e "${YELLOW}❌ Build failed${NC}"
        exit 1
    }
fi

# Start server with temp directory
echo -e "${BLUE}🚀 Starting server...${NC}"
PORT="${PORT:-3053}"
./weather \
    --port "$PORT" \
    --data "$TEST_DIR" \
    > "$TEST_DIR/server.log" 2>&1 &

SERVER_PID=$!
echo -e "Server PID: ${YELLOW}$SERVER_PID${NC}"

# Wait for server to start
echo -e "${BLUE}⏳ Waiting for server...${NC}"
for i in {1..30}; do
    if curl -s "http://localhost:$PORT/healthz" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ Server is ready!${NC}"
        break
    fi
    sleep 0.5
done

# Display information
echo ""
echo -e "${GREEN}🌤️  Server running at: ${YELLOW}http://localhost:$PORT${NC}"
echo -e "${GREEN}📊 Health check: ${YELLOW}http://localhost:$PORT/healthz${NC}"
echo -e "${GREEN}📝 API docs: ${YELLOW}http://localhost:$PORT/docs${NC}"
echo -e "${GREEN}📁 Data directory: ${YELLOW}$TEST_DIR${NC}"
echo -e "${GREEN}📋 Server log: ${YELLOW}$TEST_DIR/server.log${NC}"
echo ""
echo -e "${BLUE}Press Ctrl+C to stop the server${NC}"
echo ""

# Follow logs
tail -f "$TEST_DIR/server.log"
