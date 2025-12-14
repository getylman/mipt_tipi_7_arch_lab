FROM ubuntu:22.04

# Установка зависимостей
RUN apt-get update && apt-get install -y \
    postgresql-15 \
    postgresql-15-wal2json \
    postgresql-client-15 \
    python3 \
    python3-pip \
    python3-psycopg2 \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Установка Patroni
RUN pip3 install patroni[etcd] python-consul

# Создание пользователя для PostgreSQL
RUN groupadd -r postgres && useradd -r -g postgres postgres

# Копирование конфигурации
COPY patroni.yml /etc/patroni/

# Настройка прав
RUN chown -R postgres:postgres /etc/patroni

# Смена пользователя
USER postgres

# Рабочая директория
WORKDIR /home/postgres

# Команда запуска
CMD ["patroni", "/etc/patroni/patroni.yml"]