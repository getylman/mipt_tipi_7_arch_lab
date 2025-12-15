FROM python:3.11-slim

# Устанавливаем клиент PostgreSQL и системные зависимости
RUN apt-get update && apt-get install -y --no-install-recommends \
    postgresql-client \
    curl \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Устанавливаем Patroni
RUN pip install --no-cache-dir patroni[etcd] python-consul

# Удаляем компиляторы после установки
RUN apt-get purge -y --auto-remove gcc libpq-dev

# Создаём пользователя
RUN groupadd -r postgres && useradd -r -g postgres postgres

# Копируем конфигурацию
COPY patroni.yml /etc/patroni/
RUN chown -R postgres:postgres /etc/patroni

USER postgres
WORKDIR /home/postgres

CMD ["patroni", "/etc/patroni/patroni.yml"]