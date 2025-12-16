#!/bin/bash
set -e

echo "🧪 Запуск тестового стенда микросервиса аудита"
echo "================================================"

# Создаем результаты
mkdir -p test-results

# Останавливаем предыдущие контейнеры
docker compose -f docker-compose.test.yaml down -v 2>/dev/null || true

# Запускаем тестовый стенд
echo "1. Запуск тестовой БД PostgreSQL..."
docker compose -f docker-compose.test.yaml up -d postgres-test

echo "2. Ожидание готовности БД..."
sleep 10

echo "3. Запуск микросервиса аудита..."
docker compose -f docker-compose.test.yaml up -d audit-service-test

echo "4. Ожидание готовности микросервиса..."
sleep 15

echo "5. Проверка health-эндпоинта..."
if curl -s http://localhost:18080/health > /dev/null; then
    echo "✅ Микросервис готов!"
else
    echo "❌ Микросервис не отвечает"
    docker compose -f docker-compose.test.yaml logs audit-service-test
    exit 1
fi

echo "6. Запуск тестового клиента..."
docker compose -f docker-compose.test.yaml run --rm test-client python run_tests.py

echo ""
echo "📊 Результаты тестов сохранены в test-stub/test-results/"
ls -la test-stub/test-results/

echo ""
echo "🔍 Дополнительные команды для проверки:"
echo "   curl http://localhost:18080/stats"
echo "   curl http://localhost:18080/health"
echo "   docker compose -f docker-compose.test.yaml logs audit-service-test"
echo ""
echo "🛑 Для остановки стенда выполните: docker compose -f docker-compose.test.yaml down"
