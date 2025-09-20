# Outproxys.i2p

**Outproxys.i2p** — это проект для создания публичной базы данных аутпрокси в сети I2P.  

Аутпрокси (outproxy) — это прокси внутри сети I2P, который позволяет выходить в обычный интернет (clearnet).  

## Features
- Добавление и просмотр аутпрокси  
- Отслеживание аптайма аутпрокси  
- Простая интеграция с I2Pd  

## Build instructions

1. Установите **Go ≥ 1.25.1**  
2. В корне проекта выполните:
```bash
   go mod tidy
   go build
```

3. Отредактируйте `config.yaml` под свои нужды
4. Запустите сервер:

```bash
   ./outproxys
```

## Docker usage

Для запуска в Docker см. [README\_docker.md](README_docker.md).



