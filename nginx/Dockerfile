FROM nginx:1.25-alpine

# Установка модуля для проверки здоровья upstream
RUN apk add --no-cache nginx-mod-http-upstream-check

# Копирование конфигурации
COPY nginx.conf /etc/nginx/nginx.conf

# Создание директорий для логов
RUN mkdir -p /var/log/nginx && \
    chown -R nginx:nginx /var/log/nginx

# Смена пользователя
USER nginx

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]