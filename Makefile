.PHONY: build clean run install install-service install-config

# Цель по умолчанию
.DEFAULT_GOAL := build

# Имя бинарного файла
BINARY_NAME=kkdeamon

# Путь для установки бинарного файла
INSTALL_PATH=/usr/local/bin

# Путь к файлу конфигурации systemd
SYSTEMD_UNIT_PATH=/etc/systemd/system

# Директория, куда нужно скопировать конфигурационный файл
CONFIG_DIR=/etc/KronosKeeper

# Имя конфигурационного файла
CONFIG_NAME=kk.toml

# Цель для сборки приложения
build:
    go build -v -o $(BINARY_NAME) ./cmd/kkdeamon

# Цель для очистки временных файлов и собранного бинарного файла
clean:
    go clean
    rm -f $(BINARY_NAME)

# Цель для запуска приложения
run:
    go run ./cmd/kkdeamon/main.go

# Цель для установки бинарного файла и конфигурационного файла
install: install-service install-config
    cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)

# Цель для установки сервиса systemd
install-service:
    cp init/$(BINARY_NAME).service $(SYSTEMD_UNIT_PATH)/$(BINARY_NAME).service
    systemctl daemon-reload
    systemctl enable $(BINARY_NAME)

# Цель для копирования конфигурационного файла
install-config:
    mkdir -p $(CONFIG_DIR)
    cp ./configs/kronoskeeper_example.toml $(CONFIG_DIR)/$(CONFIG_NAME)
