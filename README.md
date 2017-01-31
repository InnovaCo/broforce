# Структура конфигурационного файла

Конфигурационный файл в формате `yaml`. Каждой задаче при запуске передается 
одноименная с задачей секция конфигурационного файла.

Пример: 
``` yaml
timeSensor:
  interval: 10

hookSensor:
  port: 8082
  url: "/github/st2"
  api-key: "123456789"

timer:
  interval: 10

```

Секция `hookSensor` будет передана задаче `hookSensor`. 
Доступ к данным осуществляется через интерфейс `ConfigData`.

Секция `logger` настраивает поведение журналирования.

Пример: 
```yaml
logger:
  file:
    name: /var/log/broforce.log
    level: debug
  fluentd:
    tag: broforce
    host: localhost
    port: 24224
    levels:
      - debug
      - info
      - warning
      - error
      - fatal
      - panic
```