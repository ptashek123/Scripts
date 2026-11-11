# Scripts

Коллекция небольших утилит на Go для повседневных задач DevOps. 
Каждая утилита — отдельный бинарник, не требует
внешних зависимостей (кроме доступа к кластеру) и собирается в одну команду.

## Утилиты

| Утилита | Назначение | README |
|---|---|---|
| [`k8s-export`](./k8s-export/) | Полный экспорт всех ресурсов кластера в чистые YAML-манифесты, пригодные для повторного `kubectl apply` | [k8s-export/README.md](./k8s-export/README.md) |

## Быстрый старт

```bash
git clone <repo-url> Scripts
cd Scripts/k8s-export

go build -o k8s-export .
./k8s-export --help
```

Или собрать бинарник для Linux без локального Go (через Docker):

```bash
docker run --rm -v "$PWD":/w -w /w/k8s-export golang:1.21 \
  sh -c "GOOS=linux GOARCH=amd64 go build -o k8s-export ."
```

## Структура репозитория

```
Scripts/
├── README.md                 # этот файл
└── k8s-export/               # утилита экспорта кластера
    ├── README.md             # подробная документация
    ├── main.go
    ├── exporter.go
    ├── cleaner.go
    ├── resources.go
    ├── go.mod
    └── go.sum
```

## Требования

- Go 1.21+
- Доступ к Kubernetes-кластеру (kubeconfig или ServiceAccount в pod)
- Права `list`/`get` на ресурсы, которые нужно выгрузить

## Философия

- **Один бинарник** — никаких внешних зависимостей в рантайме.
- **Только stdlib + официальные клиенты** — `k8s.io/client-go`, `k8s.io/apimachinery`, `sigs.k8s.io/yaml`.
- **Осторожность с кластером** — утилиты по умолчанию работают в read-only режиме и не пишут в кластер.
- **Читаемые артефакты** — YAML, папки, имена файлов: всё должно быть очевидно без открытия кода.
