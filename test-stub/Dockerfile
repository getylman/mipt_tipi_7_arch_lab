FROM python:3.11-alpine

RUN apk add --no-cache curl jq postgresql-client

WORKDIR /app

# Устанавливаем зависимости Python
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем тестовые скрипты
COPY . .

CMD ["python", "run_tests.py"]
