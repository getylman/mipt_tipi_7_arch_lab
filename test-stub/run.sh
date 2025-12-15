#!/bin/bash
set -e

# Конфигурация
TEST_ID="${1:-$(date +%Y%m%d_%H%M%S)}"
export TEST_ID
COMPOSE_FILE="docker-compose.test.yaml"

echo "🧪 Тестовый стенд микросервиса аудита"
echo "================================================"
echo "Идентификатор теста: $TEST_ID"

# Создаём директории
mkdir -p test-stub/test-results
mkdir -p test-stub/test-data

stop_services() {
    echo "Остановка тестовых сервисов..."
    docker compose -f $COMPOSE_FILE down -v --remove-orphans
}

# Обработка Ctrl+C
trap stop_services INT

# Функции
case "${2:-up}" in
    "up")
        echo "Запуск тестового стенда..."
        stop_services
        
        # Собираем и запускаем
        docker compose -f $COMPOSE_FILE build
        docker compose -f $COMPOSE_FILE up -d
        
        echo "Ожидание готовности сервисов (15 сек)..."
        sleep 15
        
        # Проверяем health
        echo "Проверка health-эндпоинтов..."
        if curl -f http://localhost:18080/health > /dev/null 2>&1; then
            echo "✅ Сервис готов к тестированию!"
            echo ""
            echo "📊 Доступные эндпоинты:"
            echo "  - Микросервис: http://localhost:18080"
            echo "  - Статистика:  http://localhost:18080/stats"
            echo "  - База данных: localhost:15433 (user: test_user, db: test_audit_db)"
            echo ""
            echo "🛠️  Команды для проверки:"
            echo "  $0 test health        # Проверить health"
            echo "  $0 test api           # Запустить API-тесты"
            echo "  $0 test load          # Нагрузочное тестирование"
            echo "  $0 test db            # Проверить БД"
            echo "  $0 logs               # Показать логи"
            echo "  $0 down               # Остановить стенд"
        else
            echo "❌ Сервис не отвечает"
            docker compose -f $COMPOSE_FILE logs audit-service-test
            exit 1
        fi
        ;;
    
    "down")
        stop_services
        ;;
    
    "logs")
        docker compose -f $COMPOSE_FILE logs -f
        ;;
    
    "test")
        case "$3" in
            "health")
                echo "Проверка health..."
                curl -s http://localhost:18080/health | jq .
                ;;
            "api")
                echo "Запуск API-тестов..."
                ./run-api-tests.sh
                ;;
            "db")
                echo "Проверка состояния БД..."
                docker exec postgres-test psql -U test_user -d test_audit_db \
                    -c "SELECT COUNT(*) as total_events FROM audit_events;"
                ;;
            "load")
                echo "Нагрузочное тестирование..."
                docker run --rm --network mipt_tipi_7_arch_lab_test-network \
                    alpine/curl:latest \
                    sh -c 'for i in $(seq 1 50); do curl -s -X POST http://audit-service-test:8080/audit/events/ \
                    -H "Content-Type: application/json" \
                    -d "{\"user\":\"load_user_$i\",\"op\":\"test_operation\"}" > /dev/null & done; wait'
                echo "Сгенерировано 50 событий"
                ;;
            *)
                echo "Доступные тесты: health, api, db, load"
                ;;
        esac
        ;;
    
    "shell")
        echo "Запуск shell в тестовом контейнере..."
        docker exec -it audit-service-test /bin/sh
        ;;
    
    "db-shell")
        echo "Запуск psql в тестовой БД..."
        docker exec -it postgres-test psql -U test_user -d test_audit_db
        ;;
    
    *)
        echo "Использование: $0 [test_id] [command]"
        echo "Команды:"
        echo "  up      - запустить тестовый стенд"
        echo "  down    - остановить стенд"
        echo "  logs    - показать логи"
        echo "  test    - запустить тесты (health|api|db|load)"
        echo "  shell   - войти в контейнер микросервиса"
        echo "  db-shell - войти в БД"
        echo ""
        echo "Пример:"
        echo "  $0 up            # Запустить стенд"
        echo "  $0 test health   # Проверить health"
        echo "  $0 test load     # Нагрузочный тест"
        echo "  $0 shell         # Войти в контейнер"
        ;;
esac
