#!/bin/bash
set -e

API_URL="http://localhost:18080"
RESULTS_DIR="test-stub/test-results"
mkdir -p $RESULTS_DIR

echo "🚀 Запуск API-тестов"
echo "================================="

run_test() {
    local name=$1
    local method=$2
    local endpoint=$3
    local data=$4
    local expected=$5
    
    echo -n "Тест '$name'... "
    
    RESPONSE=$(curl -s -w "\n%{http_code}" -X $method \
        "$API_URL$endpoint" \
        -H "Content-Type: application/json" \
        ${data:+"-d $data"})
    
    HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
    BODY=$(echo "$RESPONSE" | head -n -1)
    
    if [ "$HTTP_CODE" -eq "$expected" ]; then
        echo "✅ OK (HTTP $HTTP_CODE)"
        return 0
    else
        echo "❌ FAIL (ожидалось $expected, получили $HTTP_CODE)"
        echo "   Ответ: $BODY"
        return 1
    fi
}

# Тест 1: Health check
run_test "Health Check" GET "/health" "" 200

# Тест 2: Статистика
run_test "Stats" GET "/stats" "" 200

# Тест 3: Создание события
EVENT_DATA='{"user":"test_api_user","op":"api_test","component":"test_suite"}'
run_test "Create Event" POST "/audit/events/" "$EVENT_DATA" 201

# Тест 4: Поиск событий
run_test "Query Events" GET "/audit/events/query?ev_user=test_api_user" "" 200

# Тест 5: Валидация (должна вернуть 400)
BAD_DATA='{"user":"","op":"test"}'
run_test "Validation Error" POST "/audit/events/" "$BAD_DATA" 400

echo "================================="
echo "Тесты завершены. Результаты сохранены в $RESULTS_DIR"

# Сохраняем результаты
echo "Тесты завершены в $(date)" > "$RESULTS_DIR/last_run.txt"
