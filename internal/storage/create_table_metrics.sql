-- Создаем БД
CREATE DATABASE metrics;

-- Создаем тип метрик P.S. не работает для ТЗ. заменю на TEXT
-- CREATE TYPE mtype AS ENUM ('counter', 'gauge');

-- Это пойдет в configure
-- CREATE TABLE IF NOT EXISTS metrics (
--     id PRIMARY KEY AUTOINCREMENT,
--     mtype mtype NOT NULL;
--     mname TEXT UNIQUE NOT NULL;
--     mvalue TEXT NOT NULL;
-- )

-- DROP DATABASE metrics;