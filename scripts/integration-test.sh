#!/bin/bash
set -euo pipefail

BASE_URL="${API_URL:-http://localhost:8080}"
PROJECT_PATH="${TEST_PROJECT:-$(cd "$(dirname "$0")/../backend" && pwd)}"

echo "=== Vibe-C4 Integration Test ==="
echo "API: $BASE_URL"
echo "Project: $PROJECT_PATH"
echo ""

# 1. Health check
echo "--- Health check ---"
HEALTH=$(curl -sf "$BASE_URL/api/v1/health")
echo "$HEALTH" | jq .
if [ "$(echo "$HEALTH" | jq -r .status)" != "ok" ]; then
    echo "FAIL: Health check failed"
    exit 1
fi
echo "PASS: Health check"
echo ""

# 2. Analyze project
echo "--- Analyze project ---"
ANALYZE=$(curl -sf -X POST "$BASE_URL/api/v1/projects/analyze" \
    -H "Content-Type: application/json" \
    -d "{\"path\": \"$PROJECT_PATH\", \"name\": \"vibe-c4-backend\"}")
echo "$ANALYZE" | jq .
PROJECT_ID=$(echo "$ANALYZE" | jq -r .data.id)
STATUS=$(echo "$ANALYZE" | jq -r .data.status)
if [ "$STATUS" != "completed" ]; then
    echo "FAIL: Analysis status is $STATUS, expected completed"
    exit 1
fi
echo "PASS: Project analyzed (ID: $PROJECT_ID)"
echo ""

# 3. Get full diagram
echo "--- Get full diagram ---"
DIAGRAM=$(curl -sf "$BASE_URL/api/v1/projects/$PROJECT_ID/diagram")
SYSTEM_COUNT=$(echo "$DIAGRAM" | jq '.data.systems | length')
CONTAINER_COUNT=$(echo "$DIAGRAM" | jq '.data.containers | length')
COMPONENT_COUNT=$(echo "$DIAGRAM" | jq '.data.components | length')
echo "Systems: $SYSTEM_COUNT, Containers: $CONTAINER_COUNT, Components: $COMPONENT_COUNT"
if [ "$SYSTEM_COUNT" -lt 1 ] || [ "$CONTAINER_COUNT" -lt 1 ]; then
    echo "FAIL: Expected at least 1 system and 1 container"
    exit 1
fi
echo "PASS: Full diagram retrieved"
echo ""

# 4. Get context diagram (Level 1)
echo "--- Context diagram (Level 1) ---"
CONTEXT=$(curl -sf "$BASE_URL/api/v1/projects/$PROJECT_ID/diagram/context")
NODE_COUNT=$(echo "$CONTEXT" | jq '.data.nodes | length')
echo "Nodes: $NODE_COUNT"
if [ "$NODE_COUNT" -lt 1 ]; then
    echo "FAIL: Expected at least 1 node in context diagram"
    exit 1
fi
echo "PASS: Context diagram"
echo ""

# 5. Get container diagram (Level 2)
echo "--- Container diagram (Level 2) ---"
CONTAINERS=$(curl -sf "$BASE_URL/api/v1/projects/$PROJECT_ID/diagram/containers")
CONTAINER_NODES=$(echo "$CONTAINERS" | jq '.data.nodes | length')
CONTAINER_EDGES=$(echo "$CONTAINERS" | jq '.data.edges | length')
echo "Nodes: $CONTAINER_NODES, Edges: $CONTAINER_EDGES"
if [ "$CONTAINER_NODES" -lt 1 ]; then
    echo "FAIL: Expected at least 1 container node"
    exit 1
fi
echo "PASS: Container diagram"
echo ""

# 6. Get component diagram (Level 3) for a container
echo "--- Component diagram (Level 3) ---"
FIRST_CONTAINER_ID=$(echo "$DIAGRAM" | jq -r '.data.containers[0].id')
COMPONENTS=$(curl -sf "$BASE_URL/api/v1/projects/$PROJECT_ID/diagram/containers/$FIRST_CONTAINER_ID/components")
COMP_NODES=$(echo "$COMPONENTS" | jq '.data.nodes | length')
echo "Container: $FIRST_CONTAINER_ID, Component Nodes: $COMP_NODES"
echo "PASS: Component diagram"
echo ""

echo "=== All integration tests passed ==="
