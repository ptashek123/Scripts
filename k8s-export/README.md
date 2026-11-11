# k8s-export

Утилита для **полного экспорта ресурсов Kubernetes-кластера** в YAML-манифесты,
которые можно повторно применить в другом кластере. Полезна, когда потеряны
исходные манифесты/репозиторий, но есть живой кластер.

Каждый ресурс сохраняется в отдельный файл, из манифеста вычищаются
серверные поля (`resourceVersion`, `uid`, `managedFields`, `status` и т.д.),
а папки организуются по типу ресурса.

## Возможности

- 🔌 **Подключение любым способом** — in-cluster (ServiceAccount) →
  `--kubeconfig` → `~/.kube/config`.
- 📦 **Все стандартные ресурсы** — от `Namespace` и RBAC до `Ingress`, `HPA`,
  `PDB`, `StorageClass`, `PV/PVC` и т.д.
- 🧹 **Чистка манифестов** — удаляются серверные поля и весь `status`, чтобы
  YAML можно было сразу `kubectl apply`.
- 🗂️ **Папки по типам** — `deployments/`, `configmaps/`, `secrets/` и т.д.;
  имя файла = `metadata.name`.
- 🔒 **Секреты сохраняются как есть** — значения `data` остаются в base64,
  без маскирования и декодирования.
- 🧠 **Толерантность к ошибкам** — недоступные API, 403 Forbidden,
  отсутствующие типы ресурсов не роняют программу.
- 🎯 **Фильтрация namespace** — `-n`, `-A`, `--exclude-system`.
- 🚀 **Один бинарник**, никаких внешних зависимостей.

## Установка

### Из исходников

```bash
cd k8s-export
go mod tidy
go build -o k8s-export .
```

### Кросс-компиляция под Linux

```bash
GOOS=linux GOARCH=amd64 go build -o k8s-export .
```

## Использование

```
k8s-export [flags]
```

### Флаги

| Флаг | Сокращение | По умолчанию | Описание |
|---|---|---|---|
| `--output` | `-o` | `./k8s-export-YYYYMMDD-HHMMSS` | Папка для результата |
| `--namespace` | `-n` | — | Экспортировать только указанный namespace |
| `--all-namespaces` | `-A` | `false` | Экспортировать все namespace, включая системные |
| `--exclude-system` | | `true` | Исключать системные namespace (включается автоматически, `-A` его отключает) |
| `--kubeconfig` | | — | Путь к kubeconfig |
| `--dry-run` | | `false` | Показать, что будет сделано, без записи файлов |
| `--verbose` | `-v` | `false` | Подробный вывод |

### Примеры

```bash
# 1. Полный экспорт без системных namespace
./k8s-export

# 2. Только production, в свою папку
./k8s-export -n production -o ./production-backup

# 3. Абсолютно всё, включая kube-system
./k8s-export -A -o ./full-cluster-backup

# 4. Посмотреть, что будет выгружено, ничего не записывая
./k8s-export --dry-run

# 5. Свой kubeconfig
./k8s-export --kubeconfig /path/to/kubeconfig -o ./backup

# 6. Из пода внутри кластера (ServiceAccount)
kubectl run -it --rm k8s-export \
  --image=<registry>/k8s-export:latest \
  --restart=Never \
  --serviceaccount=k8s-export-sa \
  -- ./k8s-export -A -o /exports
```

## Структура результата

```
k8s-export-20260111-120000/
├── namespaces/
│   ├── production.yaml
│   └── staging.yaml
├── clusterroles/
│   └── my-app-role.yaml
├── clusterrolebindings/
│   └── my-app-binding.yaml
├── roles/
│   └── my-app-role.yaml
├── rolebindings/
│   └── my-app-binding.yaml
├── serviceaccounts/
│   └── my-app-sa.yaml
├── storageclasses/
│   └── standard.yaml
├── configmaps/
│   ├── app-config.yaml
│   └── nginx-config.yaml
├── secrets/
│   ├── db-credentials.yaml
│   └── tls-cert.yaml
├── persistentvolumes/
│   └── pv-data.yaml
├── persistentvolumeclaims/
│   └── pvc-data.yaml
├── networkpolicies/
│   └── deny-all.yaml
├── deployments/
│   ├── frontend.yaml
│   ├── backend.yaml
│   └── worker.yaml
├── statefulsets/
│   └── postgres.yaml
├── daemonsets/
│   └── log-collector.yaml
├── jobs/
│   └── migration.yaml
├── cronjobs/
│   └── backup.yaml
├── services/
│   ├── frontend-svc.yaml
│   └── backend-svc.yaml
├── ingresses/
│   └── main-ingress.yaml
├── ingressclasses/
│   └── nginx.yaml
├── horizontalpodautoscalers/
│   └── frontend-hpa.yaml
└── poddisruptionbudgets/
    └── backend-pdb.yaml
```

- **Имя папки** = `GVR.Resource` (например, `deployments`, `configmaps`).
- **Имя файла** = `metadata.name` + `.yaml`.
- **Пустые папки не создаются** — если в namespace нет ресурсов данного типа, ничего не пишется.
- **Один файл = один ресурс**, без multi-document `---`.

## Порядок экспорта

Ресурсы выгружаются в порядке, удобном для последующего применения:

1. **Namespaces**
2. **RBAC** — ClusterRoles, ClusterRoleBindings, Roles, RoleBindings, ServiceAccounts
3. **Storage** — StorageClasses, PersistentVolumes
4. **Config** — ConfigMaps, Secrets
5. **Storage claims** — PersistentVolumeClaims
6. **Network** — NetworkPolicies, IngressClasses
7. **Workloads** — Deployments, StatefulSets, DaemonSets, ReplicaSets, Jobs, CronJobs
8. **Services & Ingress** — Services, Endpoints, Ingresses
9. **Policy** — HPA, PDB, LimitRanges, ResourceQuotas
10. **Всё остальное** — Pods, Events, Leases, VolumeAttachments, CSIDriver, CSINode, RuntimeClass, PriorityClass

## Очистка манифестов

Из каждого объекта удаляются следующие поля:

- `metadata.resourceVersion`
- `metadata.uid`
- `metadata.selfLink`
- `metadata.creationTimestamp`
- `metadata.deletionTimestamp`
- `metadata.deletionGracePeriodSeconds`
- `metadata.finalizers`
- `metadata.ownerReferences`
- `metadata.managedFields`
- `metadata.annotations["kubectl.kubernetes.io/last-applied-configuration"]`
- весь блок `status` (удаляется полностью)

**Сохраняются**: `labels`, все остальные `annotations`, `spec`, `data` (включая
содержимое секретов), `type` и прочие «полезные» поля.

## Системные namespace

По умолчанию пропускаются:

```
kube-system, kube-public, kube-node-lease,
cert-manager, cert-manager-local,
ingress-nginx, monitoring, prometheus,
istio-system, istio-ingress, linkerd,
calico-system, tigera-operator, kube-flannel,
cattle-system, cattle-global-data,
local-path-storage, metallb-system,
longhorn-system, rook-ceph, rook-ceph-system,
opentelemetry-operator-system,
argocd, argocd-operator-system, flux-system, velero
```

Чтобы выгрузить всё, включая системные namespace, используйте `-A`
(этот флаг автоматически отключает `--exclude-system`).

## Обработка ошибок

- **403 Forbidden** — тип ресурса пропускается с предупреждением, экспорт продолжается.
- **404 NotFound / 405 MethodNotSupported** — пропускается с предупреждением.
- **Недоступные API-группы в Discovery** — пишется `[WARN]` на каждую группу, остальные выгружаются.
- **Ошибка записи файла** — предупреждение, программа продолжает работу.
- **Пустой namespace** — папка не создаётся.

Итоговый статус: если хоть что-то удалось выгрузить — программа завершится с кодом `0`.
Полный отказ (например, не удалось собрать kubeconfig) — с кодом `1`.

## Пример вывода

```
[INFO] Connected to cluster: https://10.0.0.1:6443
[INFO] Output directory: ./k8s-export-20260111-120000
[INFO] Namespaces: production, staging (skipped: kube-system, monitoring)
[STEP] Exporting cluster-scoped resources...
[INFO] ./k8s-export-20260111-120000/namespaces/production.yaml
[INFO] ./k8s-export-20260111-120000/clusterroles/my-app-role.yaml
[STEP] Exporting namespace-scoped resources...
[INFO] ./k8s-export-20260111-120000/deployments/frontend.yaml
[INFO] ./k8s-export-20260111-120000/services/frontend-svc.yaml
[INFO] ./k8s-export-20260111-120000/ingresses/main-ingress.yaml
========================================
Export complete!
Result: ./k8s-export-20260111-120000
Total files: 47
========================================
```

## Применение выгруженных манифестов

```bash
# Просто применить всё, что есть, к новому кластеру
kubectl apply -f k8s-export-20260111-120000/namespaces/
kubectl apply -f k8s-export-20260111-120000/clusterroles/
# ... и т.д. в порядке из раздела «Порядок экспорта»
```

Поскольку `status`, `managedFields` и прочие серверные поля удалены,
манифесты можно применять без предварительной чистки через `kubectl neat`
или `jq`.

## Ограничения

- **Не выгружает CRD-объекты** — только встроенные типы Kubernetes.
  Список типов задан в `resources.go`; его легко расширить.
- **Не заменяет `velero`/`kubectl get -o yaml`** — это быстрый способ
  получить читаемую файловую структуру, а не disaster-recovery-инструмент.
- **Секреты пишутся в открытом виде** (base64 — это не шифрование).
  Убедитесь, что папка с результатом не попадёт в публичный репозиторий.
- **Не выгружает встроенные объекты типа `StorageVersionMigration`,
  `APIService` и т.п.** — список расширяем, но покрывает только стандартный
  набор из ТЗ.

## Сборка и разработка

```bash
# Зависимости
go mod tidy

# Сборка
go build -o k8s-export .

# Проверка статикой
go vet ./...
```

Проект состоит из 4 файлов:

| Файл | Назначение |
|---|---|
| `main.go` | CLI-флаги, выбор kubeconfig, запуск экспорта |
| `exporter.go` | Discovery + Dynamic Client, обход ресурсов, запись YAML |
| `cleaner.go` | Удаление серверных полей из манифеста |
| `resources.go` | Список GVR-порядка и системных namespace |
