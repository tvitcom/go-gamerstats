# go-gamerstats
Some testcase

Создать api для работы с данными пользователей с использованием базы данных
MongoDb.

## Требуемый функционал

### Работа с пользователями:
- создать таблицу users и заполнить её тестовыми данными (ссылка ниже)
- разработать функционал запроса списка пользователей и информации о них
(статистика по сыгранным играм (сколько всего сыграно игр) + базовые данные)
- При выгрузке списка пользователей должна быть реализована постраничная
навигация

### Работа с играми:
- создать таблицу user_games и заполнить её тестовыми данными для каждого
пользователя (тех что добавили в таблицу users с тестового набора), набор
тестовых данных по играм можно взять по ссылке ниже на каждого
пользователя выбирается из этого набора данных рандомное количество игр
(минимум 5000 игр на пользователя).
- также должна быть возможность получить статистику сгруппированную по
номерам игр и дням
- разработать функционал получения списка рейтинга пользователей (рейтинг
считается по количеству сыгранных пользователем игр), api должно отдавать данные с
постраничной навигацией.

## Ссылки с тестовыми наборами:

[Пользователи] (https://drive.google.com/file/d/1tjubsoSwdzPK553ovvmMZs9qQwMjlKh1/view?usp=sharing)
[Данные по играм] (https://drive.google.com/file/d/1N_6pG7hxMcTJtB2MGAZZGe6_ZRfS21Mr/view?usp=sharing)

## Requirements
- golang v1.14
- mongodb
- registered domain name
- https ready connection

## Deploy
- переименовать _go.mod в go.mod
- переименовать _env в .env
- в .env указать правильные параметры подключений БД и http
- импортировать данные в mongodb скриптом import.sh
- переименовать _deploy.conf в deploy.conf
- в deploy.conf указать параметры сервера
- выполнить команду ./deploy.sh

## Test
Демо: http://yourdomain/api_v1/
Login:      guest
Password:   guest0k!

GET /api_v1/user/listing?pagenum=2
GET /api_v1/user/profile/5edad21deb7b9e2817e33c1d
GET /api_v1/user/profile/5edad21deb7b9e2817e33c1---
GET /api_v1/user/stats/5edad21deb7b9e2817e33c1d?sort=by_game&pagenum=2
GET /api_v1/user/stats/5edad21deb7b9e2817e33c1d?&sort=by_day&pagenum=1