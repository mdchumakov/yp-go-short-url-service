#!/bin/bash

# Скрипт для анализа профиля памяти с помощью pprof
# Использование: ./scripts/analyze_pprof.sh [profile_name]

PROFILE_NAME=${1:-base}
PROFILES_DIR="profiles"
PROFILE_FILE="${PROFILES_DIR}/${PROFILE_NAME}.pprof"

if [ ! -f "${PROFILE_FILE}" ]; then
    echo "Ошибка: Профиль ${PROFILE_FILE} не найден"
    echo "Сначала соберите профиль с помощью: ./scripts/collect_pprof.sh ${PROFILE_NAME}"
    exit 1
fi

echo "Анализ профиля: ${PROFILE_FILE}"
echo ""

# Показываем топ функций по использованию памяти
echo "=== Топ функций по использованию памяти ==="
go tool pprof -top "${PROFILE_FILE}" 2>/dev/null | head -20

echo ""
echo "=== Топ функций по количеству аллокаций ==="
go tool pprof -top -alloc_space "${PROFILE_FILE}" 2>/dev/null | head -20

echo ""
echo "=== Топ функций по удерживаемой памяти ==="
go tool pprof -top -inuse_space "${PROFILE_FILE}" 2>/dev/null | head -20

echo ""
echo "Для интерактивного анализа:"
echo "  go tool pprof ${PROFILE_FILE}"
echo ""
echo "Визуализация:"
echo "  go tool pprof -http=:8081 ${PROFILE_FILE}"

