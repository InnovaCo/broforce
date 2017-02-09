# broforce

Возможности:
 - прием webhook от систем: JIRA, Github, Gitlab;
 - обработка ключей consul (список серверов);
 - управление pipeline GoCD через конфигурирование;
 - обработка manifest.yml и запуск задач serve;
 - обработка комментариев JIRA и формирование сообщений slack;
 - запрос описаний issue JIRA и форммирование сообщений slack
 - прием сообщений slack и их парсинг;

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

# Ключи запуска

Список доступных ключей запуска доступен через параметр `--help`.

```
usage: broforce [<flags>]

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --config="config.yml"  Path to config.yml file.
  --show                 Show all task names.
  --allow="manifest,serve,slackSensor,hookSensor,consulSensor,outdated,gocdSheduler,jiraResolver,jiraCommenter"  
                         list of allowed tasks
  --version              Show application version.
```

Процесс может быть запущен с ключом `--allow`, в котором через `,` перечисляются задачи, 
которое разрешено запускать (по умолчанию, запускаются все доступные задачи). 

Список доступных задач выводится при использовании ключа `--show`.