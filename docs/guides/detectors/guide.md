# Гайд по разработке детекторов

## Введение

Система защиты контейнеров RuntimeRadar во время работы (runtime) обеспечивает непрерывный мониторинг событий, происходящих на уровне ядра ОС и в userspace, с помощью таких механизмов как kprobe, uprobe, tracepoint и т.д. Для получения событий через механизмы ядра Linux используется технология eBPF и Open Source продукт [cilium/tetragon](https://github.com/cilium/tetragon/). Для управления отдельными инстансами Tetragon, инжекта TracingPolicy и установки фильтров, ограничивающих поток событий, используется компонент `runtime-monitor`. Настройка инстанса Tetragon происходит без его перезапуска.

Несмотря на то, что Tetragon предоставляет свои собственные возможности для обнаружения угроз на уровне языка описания источников TracingPolicy, в решении RuntimeRadar используется совой отдельный движок для поиска угроз в событиях рантайма на основе технологии WebAssembly. За эту работу отвечает компонент `event-processor`. WebAssembly (Wasm) обеспечивает максимальную производительность детекторов, сопоставимую с нативным кодом, за счет AOT компиляции в рантайме. Одновременно с этим Wasm гарантирует полную изоляцию модулей от других модулей и встраивающего процесса. Детектор, запускаемый компонентом, не имеет никакого доступа к памяти процесса и не может ему навредить. Теоретически Wasm рантайм или компилятор могут иметь уязвимости, но они редки и сложны в эксплуатации. Одновременно со скоростью и безопасностью технология WebAssembly обеспечивает кроссплатформенность - скомпилированный `*.wasm` модуль можно запускать на любых ОС и аппаратных платформах, для которых существует соответствующий рантайм.

Детектор - это программа, написанная на языке Go, выполняющая поиск угроз в событиях рантайма. Благодаря использованию языка программирования общего назначения возможности для разработки никак не ограничены, что позволяет реализовать какую угодно по сложности логику обнаружения.

Для компиляции детекторов используется компилятор TinyGo и таргет wasip1. В директории модуля `event-processor` находится Taskfile.yml с необходимыми скриптами. Компиляция всех доступных детекторов выполняется командой
```
task detectors
```
в директории `/event-processor` репозитория продукта.

Обратите внимание, что на текущий момент времени RuntimeRadar поддерживает только детекторы, собранные компилятором TinyGo версии v0.39.0.

## Настройка окружения

Для начала работы необходимо настроить окружение и IDE. В целом может подойти любая IDE, поддерживающая разработку на языке Go, но мы рекомендуем использовать VS Code, для работы с которым в папке `/event-processor/detector/wasm/.vscode` добавлены необходимые файлы конфигурации. Для работы с другими IDE изучите содержимое файла `.vscode/settings.json` и настройте окружение аналогичным образом.

Для работы с кодом детекторов рекомендуется открыть директорию `/event-processor/detector/wasm/` в отдельном инстансе VS Code. Настройки подключатся автоматически. Если это первый запуск VS Code в контексте Go, необходимо установить рекомендуемые расширения и перезапустить IDE.

Также необходимо установить 
	- go1.25
	- TinyGo v0.39.0

Проверить что в системе установлены правильные версии тулчейна Go и TinyGo можно командой
```
tinygo version
```

## Создание нового детектора (пример)

### Разработка детектора

В качестве примера разработаем новый детектор который обнаруживает запуск `netcat` для прослушивания входящих соединений на порте 8888, что может являться признаком угрозы соответствующей горизонтальному перемещению атакующего, либо попыткой построить reverse shell. 

Для работы детектора необходимо обнаружение запуска функции ядра `inet_csk_listen_start`. Данные о TCP порте для входящих соединений содержатся в структуре `sock`, указатель на которую передается как аргумент отслеживаемой функции. Соответствующий источник (TracingPolicy) уже есть в продукте. Убедитесь, что он включен в настройках мониторинга рантайма ("Открытие сокета для входящих соединений").

В продукте доступен исходный код для ряда детекторов, работающих с событиями разных типов. Все они имеют похожую структуру. Перед началом работы стоит поискать максимально подходящий для вашего сценария детектор, чтобы использовать его в качестве основы. В нашем случае это `CS_RT_SSH_TUNNEL_USE`. 

Скопируем папку с детектором, используя его ID в качестве имени, в новую папку в той же директории что и остальные детекторы, например `TEST_NETCAT_LISTEN`.

Далее необходимо отредактировать main.go, для чего пройдем по коду сверху вниз, внося необходимые изменения.

В самом верху можно увидеть секцию с константами, содержащими метаданные детектора, включая ID, описание, автора, лицензию и т.д. 

В отредактированном виде получим:
```go
const (
	ID          = "TEST_NETCAT_LISTEN"
	Name        = "Использование netcat для входящих подключений"
	Description = "Детектор обнаруживает сетевую активность, связанную с использованием netcat для входящих подключений."
	Version     = 1
	Author      = "CS Team"
	Contact     = "email: cs@example.com"
	License     = "Apache License 2.0"
)
```

После загрузки детектора в систему эти метаданные будут отображаться в интерфейсе.

Далее объявляется секция с критериями для запуска текущего детектора - это необходимо для оптимизации работы детекторов. Точечное описание критериев позволяет запускать конкретный детектор только для конкретных типов событий и для конкретных функций, тем самым существенно сокращая суммарное время обработки события.

В критериях указываются типы событий, которые детектор должен обрабатывать, и названия функций (для тех типов, где это актуально). Подробнее с типами событий можно ознакомиться по [ссылке](https://tetragon.io/docs/reference/grpc-api/#eventtype). Детектор может работать одновременно как с разными типами событий, так и с разными функциями. Для корректной работы детектора необходимо указывать те функции, которые отслеживаются в соответствии с загруженными в систему TracingPolicy. 

Если требуется запускать детектор для всех событий, необходимо указать все типы событий, которые отслеживаются в соответствии с загруженными в систему TracingPolicy. Политики, [поставляемые](https://github.com/Runtime-Radar/runtime-radar/tree/main/runtime-monitor/pkg/model/tracingpolicy) вместе с Runtime Radar, на текущий момент работают только с `PROCESS_EXEC` и `PROCESS_KPROBE`. Если необходимо, чтобы детектор запускался для всех вызываемых функций, следует оставить список функций пустым или указать в качестве значения `*`.

Нас интересует только одна функция ядра `inet_csk_listen_start` и мы не хотим, чтобы детектор срабатывал при вызове других функций или при возникновении событий других типов. 
```go
var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"inet_csk_listen_start"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)
```

После этой секции в детекторе определяются некоторые переменные, например `sshdBin`, которая задает glob-паттерн для матчинга исполняемого файла `sshd` по имени, с учетом его полного пути. Заменим эту переменную на соответствующий шаблон для `nc`:
```go
var (
	nc = glob.MustCompile("*/nc")
)
```

Следующие участки кода можно пропустить и перейти сразу к той части, которая работает с событиями типа PROCESS_KPROBE:
```go
	switch ev := event.(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
```

В этом блоке switch-case срабатывает ветка, соответствующая интересующему нас типу события Tetragon. Неиспользуемые ветки можно удалить, или оставить, их присутствие ни на что не влияет.

Далее происходит извлечение данных из события. События Tetragon описывается через кодогенерируемые protobuf-стабы. Это позволяет работать с ними в коде модуля так же, как если бы вы работали с этими событиями вне WebAssembly, а например в каком-нибудь сервисе, взаимодействующем с Tetragon напрямую через gRPC. 

Так же как и в случае с обычным protobuf, рекомендуется использовать геттеры там где это возможно, чтобы избежать паники при случайном обращении к `nil`.

В следующем блоке кода из события извлекаются необходимые нам данные:
- Путь к исполняемому файлу (`binary`)
- Имя функции (`function`)
	
```go
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		function := kprobe.GetFunctionName()
```

Далее используем новый шаблон `nc`, обратившись к методу `Match` для проверки матчинга полученного пути к исполняемому файлу.
```go
		if !(nc.Match(binary) && function == "inet_csk_listen_start") {
			return resp, nil
		}
```

Одновременно с этим выполняется проверка что целевая функция (`inet_csk_listen_start`) правильная. Если это не так детектор завершается с нулевым результатом, то есть без обнаруженной угрозы.

Теперь нам осталось только проверить что порт который прослушивает `netcat` соответствует целевому (8888). Перепишем оставшиеся секции следующим образом
```go
		args := kprobe.GetArgs()
		if len(args) < 1 {
			return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
		}

		sport := args[0].GetSockArg().GetSport()
		if sport == 8888 {
			resp.Severity = api.DetectResp_CRITICAL
			return resp, nil
		}
```

Первым делом мы извлекаем слайс указателей на объекты аргументов (`args`). Обратите внимание, что сразу после этого проверяется, что он имеет необходимую минимальную длину. Если это не так, детектор возвращает пустой результат и ошибку, содержащую подробное сообщение с помощью функции стандартной библиотеки Go `fmt.Errorf`. Рекомендуется добавлять такие проверки во всех местах, где предполагается обращение по индексу слайса, чтобы избежать паники при выполнении кода. 

Паника внутри конкретного детектора никак не повлияет на другие детекторы или компонент `event-processor`, они продолжат работать. Тем не менее такая ошибка может привести к неработоспособности детектора.

Поскольку мы точно знаем, что необходимый нам объект является первым, мы можем извлечь его по индексу (0), и, используя цепочку `GetSockArg().GetSport()`, извлекаем номер порта.

Если целевой порт имеет значение 8888, мы выставляем Severity угрозы в значение Critical и таким образом возвращаем результат, соответствующий найденной угрозе.

Полный исходный код получившегося детектора представлен в конце.

### Компиляция детектора

Для компиляции в системе должен быть установлен компилятор TinyGo рекомендованной версии и утилита `task`:
- https://tinygo.org/getting-started/install/
- https://taskfile.dev/docs/installation

Если детектор был добавлен в директорию `/event-processor/detector/wasm/` в соответствии с описанием выше, то рекомендуемый способ компиляции это запуск сценария компиляции через `task`:
```
task detectors
```
В результате `*.wasm` файл нового детектора появится в папке `/event-processor/deploy/`.


Альтернативный вариант сборки детектора это запуск TinyGo вручную, например из директории нового детектора `/event-processor/detector/wasm/TEST_NETCAT_LISTEN`:
```
tinygo build -target=wasip1 -scheduler=none --no-debug -o TEST_NETCAT_LISTEN.wasm
```
 
### Добавление детектора в систему

Для того чтобы добавить детектор в систему, перейдите на вкладку "Детекторы" в разделе "Рантайм".
В открывшемся окне нужно выбрать `*.wasm` файл детектора (например, через drag and drop) и нажать на кнопку "Добавить". 

Детектор добавляется без перезапуска каких-либо компонентов, изменения вступают в силу на всех репликах `event-processor` в течение минуты.

### Проверка работы 

Для проверки лучше всего использовать тестовый под, мониторинг которого был настроен в системе:
1. Откройте шелл в контейнер наблюдаемого пода, используя `kubectl` или любой другой иснтрумент
2. Установите `netcat`
3. Выполните команду
```
nc -vv -l -p 8888
```
4. Рекомендуется выполнить еще несколько простых команд, например `ls -al` чтобы буфер накопленных сообщений сохранился в БД
5. Перейдите на вкладку "События" раздела "Рантайм", и включите фильтр "События только с угрозами". В списке событий должны появиться события, в которых детекторы обнаружили какие-либо угрозы, включая новый детектор TEST_NETCAT_LISTEN

### Исходный код детектора

Полный код детектора представлен здесь.

```go
package main

import (
	"context"
	"fmt"

	"github.com/gobwas/glob"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api"
	"github.com/runtime-radar/runtime-radar/event-processor/detector/api/tetragon"
)

const (
	ID          = "TEST_NETCAT_LISTEN"
	Name        = "Использование netcat для входящих подключений"
	Description = "Детектор обнаруживает сетевую активность, связанную с использованием netcat для входящих подключений."
	Version     = 1
	Author      = "CS Team"
	Contact     = "email: cs@example.com"
	License     = "Apache License 2.0"
)

var (
	// triggerCriteria sets Trigger Criteria as map of events types to corresponding functions
	// which will be used by Detector. If function names are not applicable for
	// a particular event type, such as "PROCESS_EXEC", leave slice empty or use
	// wildcard "*".
	triggerCriteria = map[string][]string{
		"PROCESS_KPROBE": {"inet_csk_listen_start"},

		// Examples:
		//
		// "PROCESS_KPROBE": {"security_file_permission", "security_mmap_file", "security_path_truncate"},
		// In order to process all possible functions leave right-hand part empty or use wildcard "*":
		// "PROCESS_EXEC": {},
		// same as:
		// "PROCESS_EXEC": {"*"},
	}
)

var (
	nc = glob.MustCompile("*/nc")
)

// main is required for TinyGo to compile to Wasm.
func main() {
	api.RegisterDetector(Detector{})
}

type Detector struct{}

func (d Detector) Info(ctx context.Context, req *api.InfoReq) (*api.InfoResp, error) {
	return &api.InfoResp{
		Id:          ID,
		Name:        Name,
		Description: Description,
		Version:     Version,
		Author:      Author,
		Contact:     Contact,
		License:     License,
	}, nil
}

func (d Detector) Detect(ctx context.Context, req *api.DetectReq) (*api.DetectResp, error) {
	// Detector info added to DetectResp because detector info is always correlated to response, thus
	// to avoid +1 Wasm call on detect.
	resp := &api.DetectResp{
		Id:          ID,
		Name:        Name,
		Description: Description,
		Version:     Version,
		Author:      Author,
		Contact:     Contact,

		// Default response indicates that nothing detected (this is redundant and put here just for reference,
		// as Severity == api.DetectResp_NONE == 0 when omitted (default zero value)).
		Severity: api.DetectResp_NONE,
	}

	event := req.GetEvent().GetEvent()

	switch ev := event.(type) {
	case *tetragon.GetEventsResponse_ProcessExec:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessExit:
		// Nothing here
	case *tetragon.GetEventsResponse_ProcessKprobe:
		kprobe := ev.ProcessKprobe
		binary := kprobe.GetProcess().GetBinary()
		function := kprobe.GetFunctionName()

		if !(nc.Match(binary) && function == "inet_csk_listen_start") {
			return resp, nil
		}

		args := kprobe.GetArgs()
		if len(args) < 1 {
			return nil, fmt.Errorf("unexpected args len, got %d, want >= 1", len(args))
		}

		sport := args[0].GetSockArg().GetSport()
		if sport == 8888 {
			resp.Severity = api.DetectResp_CRITICAL
			return resp, nil
		}

		// Nothing here
	case *tetragon.GetEventsResponse_ProcessTracepoint:
		// Nothing here
	}

	return resp, nil
}

func (d Detector) TriggerCriteria(ctx context.Context, req *api.TriggerCriteriaReq) (*api.TriggerCriteriaResp, error) {
	resp := &api.TriggerCriteriaResp{
		Criteria: make(map[string]*api.TriggerCriteriaResp_FuncNames, len(triggerCriteria)),
	}

	for k, v := range triggerCriteria {
		resp.Criteria[k] = &api.TriggerCriteriaResp_FuncNames{FuncNames: v}
	}

	return resp, nil
}
```
