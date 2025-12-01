#!/bin/bash

# Скрипт для сбора профиля памяти с помощью pprof
# Использование: ./scripts/collect_pprof.sh [profile_name]

PROFILE_NAME=${1:-base}
PROFILES_DIR="profiles"
SERVER_URL="http://localhost:8080"
PROFILE_FILE="${PROFILES_DIR}/${PROFILE_NAME}.pprof"

echo "Сбор профиля памяти..."
echo "Имя профиля: ${PROFILE_NAME}"
echo "Файл профиля: ${PROFILE_FILE}"

# Проверяем, запущен ли сервер
if ! curl -s "${SERVER_URL}/ping" > /dev/null; then
    echo "Ошибка: Сервер не запущен на ${SERVER_URL}"
    echo "Пожалуйста, запустите сервер перед сбором профиля"
    exit 1
fi

# Создаем директорию profiles, если её нет
mkdir -p "${PROFILES_DIR}"

# Собираем профиль памяти (heap)
echo "Собираем профиль памяти (heap)..."
curl -s "${SERVER_URL}/debug/pprof/heap" > "${PROFILE_FILE}"

if [ $? -eq 0 ]; then
    echo "Профиль успешно сохранен в ${PROFILE_FILE}"
    echo "Размер файла: $(du -h ${PROFILE_FILE} | cut -f1)"
else
    echo "Ошибка при сборе профиля"
    exit 1
fi

echo ""
echo "Для анализа профиля:"
echo "  go tool pprof ${PROFILE_FILE}"
echo ""
echo "Полезные команды pprof:"
echo "  top        - показать топ функций по использованию памяти"
echo "  list <func> - показать исходный код функции"
echo "  web        - открыть интерактивный граф (требует graphviz)"
echo "  peek <func> - показать вызывающие функции"

