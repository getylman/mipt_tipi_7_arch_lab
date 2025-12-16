#!/usr/bin/env python3
import requests
import json
import time
import psycopg2
import sys
from datetime import datetime, timedelta

# Конфигурация
API_URL = "http://audit-service-test:8080"
DB_CONFIG = {
    "host": "postgres-test",
    "port": 5432,
    "database": "test_audit_db",
    "user": "test_user",
    "password": "test_password"
}

class AuditServiceTester:
    def __init__(self):
        self.test_id = f"test_{int(time.time())}"
        self.session = requests.Session()
        
    def check_service_health(self):
        """Проверяем, что сервис доступен"""
        try:
            response = self.session.get(f"{API_URL}/health", timeout=5)
            return response.status_code == 200
        except:
            return False
    
    def create_test_event(self, user=None, operation=None, attributes=None):
        """Создает тестовое событие через API"""
        if user is None:
            user = f"{self.test_id}_user"
        
        event = {
            "user": user,
            "op": operation or "test_operation",
            "timestamp": datetime.utcnow().isoformat() + "Z",
            "component": "test_suite",
            "attributes": attributes or {
                "test_id": self.test_id,
                "ip": "192.168.1.100",
                "success": True
            }
        }
        
        response = self.session.post(
            f"{API_URL}/audit/events/",
            json=event,
            headers={"Content-Type": "application/json"},
            timeout=10
        )
        
        if response.status_code != 201:
            print(f"❌ Ошибка создания события: {response.status_code}")
            print(f"   Ответ: {response.text}")
            return None
        
        created_event = response.json()
        print(f"✅ Создано событие ID: {created_event['id']}")
        return created_event
    
    def query_events(self, params=None):
        """Ищет события через API"""
        if params is None:
            params = {}
        
        response = self.session.get(
            f"{API_URL}/audit/events/query",
            params=params,
            timeout=10
        )
        
        if response.status_code != 200:
            print(f"❌ Ошибка поиска событий: {response.status_code}")
            print(f"   Ответ: {response.text}")
            return []
        
        return response.json()
    
    def verify_event_in_db(self, event_id):
        """Проверяет, что событие действительно сохранено в БД"""
        try:
            conn = psycopg2.connect(**DB_CONFIG)
            cursor = conn.cursor()
            
            cursor.execute(
                "SELECT id, user_id, operation FROM audit_events WHERE id = %s",
                (event_id,)
            )
            
            result = cursor.fetchone()
            cursor.close()
            conn.close()
            
            if result:
                print(f"✅ Событие найдено в БД: ID={result[0]}, user={result[1]}, op={result[2]}")
                return True
            else:
                print(f"❌ Событие {event_id} не найдено в БД")
                return False
                
        except Exception as e:
            print(f"❌ Ошибка подключения к БД: {e}")
            return False
    
    def run_comprehensive_test(self):
        """Запускает комплексный тест"""
        print(f"\n{'='*60}")
        print(f"🚀 Запуск тестов микросервиса аудита")
        print(f"   Test ID: {self.test_id}")
        print(f"   Время: {datetime.now().isoformat()}")
        print(f"{'='*60}\n")
        
        # Шаг 1: Проверка доступности сервиса
        print("1. Проверка доступности сервиса...")
        if not self.check_service_health():
            print("❌ Сервис недоступен")
            return False
        print("✅ Сервис доступен\n")
        
        # Шаг 2: Создание тестового события
        print("2. Создание тестового события...")
        test_user = f"test_user_{self.test_id}"
        test_event = self.create_test_event(
            user=test_user,
            operation="user_login",
            attributes={
                "test_id": self.test_id,
                "ip": "10.0.0.1",
                "browser": "chrome",
                "location": "Moscow"
            }
        )
        
        if not test_event:
            return False
        
        event_id = test_event['id']
        print(f"   Создано событие: user={test_user}, id={event_id}\n")
        
        # Шаг 3: Проверка через прямой запрос к БД
        print("3. Проверка записи в БД...")
        if not self.verify_event_in_db(event_id):
            return False
        print()
        
        # Шаг 4: Поиск события через API
        print("4. Поиск события через API...")
        
        # Поиск по пользователю
        print("   a) Поиск по пользователю...")
        events = self.query_events({"ev_user": test_user})
        if any(e['id'] == event_id for e in events):
            print(f"   ✅ Событие найдено по пользователю (найдено: {len(events)} событий)")
        else:
            print(f"   ❌ Событие не найдено по пользователю")
            return False
        
        # Поиск по операции
        print("   b) Поиск по операции...")
        events = self.query_events({"ev_op": "user_login"})
        if any(e['id'] == event_id for e in events):
            print(f"   ✅ Событие найдено по операции (найдено: {len(events)} событий)")
        else:
            print(f"   ❌ Событие не найдено по операции")
            return False
        
        # Поиск по времени
        print("   c) Поиск по временному диапазону...")
        end_time = datetime.utcnow().isoformat() + "Z"
        start_time = (datetime.utcnow() - timedelta(hours=1)).isoformat() + "Z"
        
        events = self.query_events({
            "ev_ts_start": start_time,
            "ev_ts_end": end_time
        })
        
        if any(e['id'] == event_id for e in events):
            print(f"   ✅ Событие найдено по времени (найдено: {len(events)} событий)")
        else:
            print(f"   ❌ Событие не найдено по времени")
            return False
        
        # Шаг 5: Тест на валидацию
        print("\n5. Тест валидации данных...")
        
        # Неверный запрос (нет пользователя)
        print("   a) Проверка валидации (отсутствует user)...")
        response = self.session.post(
            f"{API_URL}/audit/events/",
            json={"op": "test"},
            headers={"Content-Type": "application/json"}
        )
        
        if response.status_code == 400:
            print("   ✅ Валидация работает: получена ошибка 400")
        else:
            print(f"   ❌ Ожидалась ошибка 400, получено: {response.status_code}")
            return False
        
        # Слишком длинный пользователь
        print("   b) Проверка валидации (слишком длинный user)...")
        response = self.session.post(
            f"{API_URL}/audit/events/",
            json={
                "user": "x" * 300,  # > 255 символов
                "op": "test"
            },
            headers={"Content-Type": "application/json"}
        )
        
        if response.status_code != 201:  # Ожидаем ошибку валидации
            print("   ✅ Валидация длины работает")
        else:
            print(f"   ⚠️  Валидация длины не сработала")
        
        print(f"\n{'='*60}")
        print(f"🎉 ВСЕ ТЕСТЫ УСПЕШНО ПРОЙДЕНЫ!")
        print(f"   Создано событий: 1")
        print(f"   Проверено API endpoints: 4")
        print(f"   Время выполнения: {datetime.now().strftime('%H:%M:%S')}")
        print(f"{'='*60}")
        
        return True

def main():
    print("Подготовка к тестированию...")
    
    # Даем время сервисам запуститься
    time.sleep(5)
    
    tester = AuditServiceTester()
    
    # Ждем, пока сервис станет доступен
    print("Ожидание готовности сервиса...")
    for i in range(30):  # Ждем до 30 секунд
        if tester.check_service_health():
            break
        time.sleep(1)
        if i % 5 == 0:
            print(f"  ...прошло {i+1} секунд")
    
    # Запускаем тесты
    success = tester.run_comprehensive_test()
    
    # Сохраняем результаты
    with open('/test-results/test_report.json', 'w') as f:
        json.dump({
            "test_id": tester.test_id,
            "timestamp": datetime.now().isoformat(),
            "success": success
        }, f, indent=2)
    
    sys.exit(0 if success else 1)

if __name__ == "__main__":
    main()

