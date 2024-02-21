# KronosKeeper: Управление резервными копиями

**Внимание:** Проект находится на стадии глубокой разработки.

KronosKeeper - это приложение для создания резервных копий. Основные возможности:

- Создание резервных копий на основе архивов.
- Выгрузка резервных копий на Google Cloud.
- Запуск задач резервного копирования по расписанию.
- Уведомление в Telegram о результатах резервного копирования.

## Установка

В проекте используеться go 1.21.6. Веройтно перед сборкой программы вам нужно будет его установить.
Используйте для сборки и установки:

```bash
make
make install
```

Конфигурационный файл лежит в /etc/KronosKeeper/kk.toml

Для управления демоном используйте:

```bash
systemctl start kkdeamon
systemctl status kkdeamon
```

## Использование

Для работы с Google Cloud требуется файл `credentials.json`, который можно создать на [сайте Google Cloud](https://cloud.google.com).

Для работы с Telegram необходим токен (получаем у BotFather) и ваш ID, который можно узнать у [userinfobot](https://t.me/userinfobot).

Формат сжатия: только zip.

### Поддерживаемые форматы конфигурационного файла

```toml
log_level = "DEBUG"  # Уровень журналирования 
log_path = ""        # Путь к файлу журнала configs/kronoskeeper.log

### Настройка уведомлений
[telegram]
token = ""           # API ключ Telegram (введите ваш собственный ключ)
chat_id = ""         # ID чата в телеграм, используйте userinfobot

### Настройка удаленных хранилищ данных
[storage]
[storage.gCloud]
credentials_json = "configs/credentials.json"  # Путь к JSON-файлу с учетными данными для Google Cloud

### Настройка юнитов/задач бекапов
[[Unit]]
name = "nginx"        # Имя юнита/задачи бэкапа
crontabTask = "* * * * * *"  # Расписание Cron 
input = ["/tmp/test"]  # Пути для бэкапа
output = "/tmp"        # Путь для сохранения бэкапов
compressFormat = "zip"  # Формат сжатия бэкапов
compressExclude = ["file1", "*.zip"]  # Исключения из сжатия
remotestorages = ["gCloud"]  # Список удаленных хранилищ, куда отправлять бэкапы
remoteDir = "hostnamemyserver"  # Папка на удаленном хранилище для сохранения бэкапов

### Структура папок для каждого юнита бекапа:
Для каждого юнита бекапа создается папка с его именем. В этой папке создаются подпапки с названием ГОД-МЕСЯЦ, а в них сохраняются архивы с именем в формате ДЕНЬ-ЧАСЫ:МИНУТЫ-name.
```bash

Structure Dir nginx
    |-- 2024-03/
        |-- 03-13:37-nginx.zip
        |-- 02-13:37-nginx.zip
    |-- 2024-02/
        |-- 21-13:37-nginx.zip
        |-- 20-13:37-nginx.zip

```
