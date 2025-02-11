# TransactionSystem

TransactionSystem — это система для управления транзакциями и кошельками, позволяющая выполнять переводы средств между кошельками, получать историю транзакций, управлять балансом и осуществлять другие операции.

## Основные возможности

- **Перевод средств:** Реализация перевода денег с одного кошелька на другой с проверкой корректности транзакции.
- **Управление кошельками:** Создание, удаление и получение информации о кошельках, включая баланс.
- **История транзакций:** Получение списка последних транзакций с возможностью указания количества возвращаемых записей.
- **Получение транзакций по параметрам:** Поиск транзакции по идентификатору, или по отправителю, получателю и времени создания.

## Структура проекта

- **/api**  
  Содержит обработчики HTTP-запросов (эндпоинты API), реализованные с помощью [Gorilla Mux](https://github.com/gorilla/mux).
  

	/project-root
	│── /cmd/server/main.go         # Точка входа в приложение

	│── /config                     # Конфигурация приложения

	│── /internal                   # Внутренний код приложения

	│   ├── /database               # Подключение к базе данных/Запуск миграций

	│   ├── /models                 # Определения структур данных

	│   ├── /repository             # Работа с БД (хранение данных)

	│   ├── /service                # Бизнес-логика

	│── /api                        # Обработчики HTTP-запросов

	│── /docs                       # Документация

	│── /test                       # Тестирование

	│── /build                      # Сборка Docker образа, запуск проекта через docker-compose

	│── go.mod

	│── go.sum

 Чтобы запустить любой из тестов надо перенести файл в TransactionSystem/ и запустить - 
  
        go test file_test.go
 
 Тесты покрывают слой бизнес логики. Тесты не информативны, так как писались исключительно в целях проверить текущие результаты и убедиться в работоспособности проекта.

## API Документация

Для детального описания эндпоинтов API, параметров запросов и формата ответов смотрите файл [API.md](./api/API.md).

### Пример эндпоинта: Перевод средств

- **URL:** `/api/send`
- **Метод:** `POST`
- **Тело запроса:**
  ```json
  {
    "from": "wallet_123",
    "to": "wallet_456",
    "amount": 100.50
  }

## Технологии

    Язык: Go (Golang)
    Маршрутизация: Gorilla Mux
    Работа с JSON/YAML: encoding/json, gopkg.in/yaml.v2
    Логирование: стандартный пакет log
    База данных: PostgreSQL, pgx

## Архитектура

Проект частично реализует чистую архитектуру:

    API слой: Обработка HTTP-запросов, валидация входных данных и вызов соответствующих сервисов.
    Сервисный слой: Реализация бизнес-логики приложения.
    Слой доступа к данным (Repository): Выполнение операций с базой данных, таких как запросы, обновления, вставки и удаления данных.

## Заметки

В ./api присутствует два файла [handlers.go](api%2Fhandlers.go) и [handlers_main.go](api%2Fhandlers_main.go). Тот, что с припиской main, содержит в себе эндпоинты требуемые в ТЗ. Всё что содержиться в другом файле это эндопинты, которые мне показались необходимыми для поставленной задачи.

Касательно частичного преноса бизнес логики перевода денег в слой данных, метод - 

    ExecuteTransfer(ctx context.Context, from, to string, balance_from, balance_to, amount float64) error 

Метод не выполняет CRUD операцию и затрагивает сразу две сущности Wallet и Tranaction. Передо мной стояла задача - в случае ошибки вернуть систему в исходное состояние. Обернуть всё в одну транзакцию БД оказалось проще, чем 'руками' восстанавливать балансы, что могло бы привести к появлению уязвимостей. 
Я решил отойти от чистой архитектуры, чтобы избежать ненужных на мой взгляд уязвимостей.

Dockerfile и docker-compose не дописаны! Так же и с логированием и ответами сервера.
