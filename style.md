## Стиль

### **Никита Лабораторная работа 3**

### Группировать похожие

Go поддерживает группировку схожих объявлений.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import "a"
import "b"
```

</td><td>

```go
import (
  "a"
  "b"
)
```

</td></tr>
</tbody></table>

Это также относится к константам, переменным и объявлениям типов.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go

const a = 1
const b = 2



var a = 1
var b = 2



type Area float64
type Volume float64
```

</td><td>

```go
const (
  a = 1
  b = 2
)

var (
  a = 1
  b = 2
)

type (
  Area float64
  Volume float64
)
```

</td></tr>
</tbody></table>

Группируйте только связанные между собой объявления. Не группируйте несвязанные между собой объявления.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
type Operation int

const (
  Add Operation = iota + 1
  Subtract
  Multiply
  EnvVar = "MY_ENV"
)
```

</td><td>

```go
type Operation int

const (
  Add Operation = iota + 1
  Subtract
  Multiply
)

const EnvVar = "MY_ENV"
```

</td></tr>
</tbody></table>

Группы не ограничены в применении. Например, их можно использовать внутри функций.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
func f() string {
  red := color.New(0xff0000)
  green := color.New(0x00ff00)
  blue := color.New(0x0000ff)

  // ...
}
```

</td><td>

```go
func f() string {
  var (
    red   = color.New(0xff0000)
    green = color.New(0x00ff00)
    blue  = color.New(0x0000ff)
  )

  // ...
}
```

</td></tr>
</tbody></table>

Исключение: объявления переменных, особенно внутри функций, следует группировать, если они объявлены рядом с другими переменными. Делайте это для переменных, объявленных вместе, даже если они не связаны между собой.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
func (c *client) request() {
  caller := c.name
  format := "json"
  timeout := 5*time.Second
  var err error

  // ...
}
```

</td><td>

```go
func (c *client) request() {
  var (
    caller  = c.name
    format  = "json"
    timeout = 5*time.Second
    err error
  )

  // ...
}
```

</td></tr>
</tbody></table>

### Порядок импортов

Должно быть две группы импорта:

Стандартная библиотека
Все остальное
Это группировка, применяемая goimports по умолчанию.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import (
  "fmt"
  "os"
  "go.uber.org/atomic"
  "golang.org/x/sync/errgroup"
)
```

</td><td>

```go
import (
  "fmt"
  "os"

  "go.uber.org/atomic"
  "golang.org/x/sync/errgroup"
)
```

</td></tr>
</tbody></table>

### Имена пакетов

При именовании пакетов выбирайте имя, которое:

- Все строчные буквы. Без заглавных букв и подчёркиваний.
- Не требует переименования с использованием именованных импортов в большинстве мест вызова.
- Коротко и ёмко. Помните, что имя указывается полностью при каждом вызове. 
- Не во множественном числе. Например, net/url, не net/urls.
- Не «common», «util», «shared» или «lib». Это плохие, неинформативные названия.


### Импорт псевдонимов

Если имя пакета не совпадает с последним элементом пути импорта, необходимо использовать псевдоним импорта.

```go
import (
  "net/http"

  client "example.com/client-go"
  trace "example.com/trace/v2"
)
```

Во всех остальных случаях следует избегать использования импортных псевдонимов, если только нет прямого конфликта между импортируемыми данными.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
import (
  "fmt"
  "os"
  runtimetrace "runtime/trace"

  nettrace "golang.net/x/trace"
)
```

</td><td>

```go
import (
  "fmt"
  "os"
  "runtime/trace"

  nettrace "golang.net/x/trace"
)
```

</td></tr>
</tbody></table>

### Группировка и упорядочивание функций

- Функции следует сортировать в приблизительном порядке вызова.
- Функции в файле должны быть сгруппированы по приемнику.

Поэтому экспортируемые функции должны располагаться в файле первыми struct после constопределений var.

Символ newXYZ()/ NewXYZ()может появляться после определения типа, но перед остальными методами приемника.

Поскольку функции сгруппированы по получателю, простые служебные функции должны располагаться ближе к концу файла.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
func (s *something) Cost() {
  return calcCost(s.weights)
}

type something struct{ ... }

func calcCost(n []int) int {...}

func (s *something) Stop() {...}

func newSomething() *something {
    return &something{}
}
```

</td><td>

```go
type something struct{ ... }

func newSomething() *something {
    return &something{}
}

func (s *something) Cost() {
  return calcCost(s.weights)
}

func (s *something) Stop() {...}

func calcCost(n []int) int {...}
```

</td></tr>
</tbody></table>

### Уменьшение вложенности

По возможности следует уменьшить вложенность кода, обрабатывая ошибки/особые условия в первую очередь и возвращая управление на раннем этапе или продолжая цикл. Сократите объём кода с несколькими уровнями вложенности.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
for _, v := range data {
  if v.F1 == 1 {
    v = process(v)
    if err := v.Call(); err == nil {
      v.Send()
    } else {
      return err
    }
  } else {
    log.Printf("Invalid v: %v", v)
  }
}
```

</td><td>

```go
for _, v := range data {
  if v.F1 != 1 {
    log.Printf("Invalid v: %v", v)
    continue
  }

  v = process(v)
  if err := v.Call(); err != nil {
    return err
  }
  v.Send()
}
```

</td></tr>
</tbody></table>

### Ненужное Else

Если переменная задана в обеих ветвях if, ее можно заменить одним if.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
var a int
if b {
  a = 100
} else {
  a = 10
}
```

</td><td>

```go
a := 10
if b {
  a = 100
}
```

</td></tr>
</tbody></table>

### Встраивание в структуры

Встроенные типы должны располагаться в верхней части списка полей структуры, а встроенные поля должны быть отделены от обычных полей пустой строкой.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
type Client struct {
  version int
  http.Client
}
```

</td><td>

```go
type Client struct {
  http.Client

  version int
}
```

</td></tr>
</tbody></table>

Встраивание должно обеспечивать ощутимые преимущества, например, добавление или расширение функциональности семантически обоснованным образом. При этом не должно возникать никаких негативных последствий для пользователя 

Встраивание не должно :

- Иметь исключительно косметический или ориентированный на удобство характер.
- Сделать внешние типы более сложными в создании и использовании.
- Влиять на нулевые значения внешних типов. Если внешний тип имеет полезное нулевое значение, он должен иметь полезное нулевое значение и после внедрения внутреннего типа.
- Изменять API внешнего типа или семантику типа.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
type Book struct {
    io.ReadWriter
}

var b Book
b.Read(...)  // panic: nil pointer
b.String()   // panic: nil pointer
b.Write(...) // panic: nil pointer
```

</td><td>

```go
type Book struct {
    bytes.Buffer
}

var b Book
b.Read(...)  // ok
b.String()   // ok
b.Write(...) // ok
```

</td></tr>
<tr><td>

```go
type Client struct {
    sync.Mutex
    sync.WaitGroup
    bytes.Buffer
    url.URL
}
```

</td><td>

```go
type Client struct {
    mtx sync.Mutex
    wg  sync.WaitGroup
    buf bytes.Buffer
    url url.URL
}
```

</td></tr>
</tbody></table>

### Использование логических операторов

Логическое И (&&) и ИЛИ (||): Объединяйте несколько условий в одно, чтобы избежать вложенных конструкций if.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
if age > 18 {
    if hasLicense {
        // ...
    }
}

```
</td><td>

```go
if age > 18 && hasLicense {
    // ...
}
```
</td></tr>
</tbody></table>

Отрицание (!): Используйте для проверки инвертированных состояний.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
if someCondition == false { // Менее читаемо
    // ...
}


```
</td><td>

```go
if !someCondition {
    // ...
}
```
</td></tr>
</tbody></table>

Используйте switch для случаев, где есть одна переменная или выражение, которое нужно сравнить с несколькими возможными значениями. Это более читаемо, чем длинные цепочки if-else if-else if.

<table>
<thead><tr><th>Bad</th><th>Good</th></tr></thead>
<tbody>
<tr><td>

```go
if command == "start" {
    // ...
} else if command == "stop" {
    // ...
} else if command == "restart" {
    // ...
}

```
</td><td>

```go
switch command {
case "start":
    // ...
case "stop":
    // ...
case "restart":
    // ...
default:
    // ...
}

```
</td></tr>
</tbody></table>


### DRY и KISS

**DRY (Don't Repeat Yourself)**

**Суть:** Каждая часть информации в системе должна иметь единственное, непротиворечивое и авторитетное представление.

**Цель:** Избежать дублирования кода, что упрощает внесение изменений и снижает вероятность ошибок. Если нужно изменить что-то, правка производится в одном месте, а не в нескольких.

**Примеры реализации:**
- Вынесение повторяющегося кода в отдельные функции или методы.
- Использование абстракций, классов, интерфейсов и шаблонов для общей функциональности.
- Использование готовых библиотек и фреймворков вместо написания собственного кода для стандартных задач. 


**KISS (Keep It Simple, Stupid)**

**Суть:** Избегайте излишней сложности. Система должна быть максимально простой.

**Цель:** Сделать код более читаемым, поддерживаемым и простым в модификации.

**Примеры реализации:**
- Методы должны быть короткими и решать одну конкретную задачу.
- Использовать понятные и самодокументирующиеся имена переменных и методов.
- Разделять код на логически связанные части (модули, классы) для лучшей поддержки.
- Не добавлять функционал «на всякий случай», если он не требуется прямо сейчас.

### Работа с базами данных и обработка запросов к ним. 

Использовать стандартный пакет `database/sql`

**Подключение к базе данных:**
- Использовать `sql.Open()` для открытия соединения с базой данных, указав драйвер и строку подключения.
- Не забудьте закрыть соединение с помощью `defer db.Close()`.
- Пример: 
```go
import "database/sql"

db, err := sql.Open("postgres", "user=pqgotest dbname=pqgotest sslmode=disable")
if err != nil {
    // обработка ошибки
}
defer db.Close()
```

**Выполнение запроса:**

- Для выборки данных использовать `db.Query()` и обработать `*sql.Rows` с помощью `rows.Next()` и `rows.Scan()`.

```go
rows, err := db.Query("SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    // обработка ошибки
}
defer rows.Close()

for rows.Next() {
    var id int
    var name string
    err := rows.Scan(&id, &name)
    // обработка ошибки и данных
}

```

- Для операций, меняющих данные (INSERT, UPDATE, DELETE): использовать `db.Exec()`

```go
result, err := db.Exec("INSERT INTO users (name) VALUES ($1)", "Alice")
```

**Общие рекомендации:**
- **Безопасность:** Всегда использовать подготовленные выражения (placeholder'ы) с параметрами, а не конкатенацию строк, чтобы защититься от SQL-инъекций.
- **Обработка ошибок:** Всегда проверять ошибки, которые возвращают функции работы с БД.
- **Закрытие соединений:** Не забывать закрывать `*sql.Rows` после использования с помощью `defer rows.Close()`

### **Давыдов Денис Лабораторная работа 3**

## Обработка ошибок

### Обрабатывайте ошибки только один раз

Каждая ошибка должна быть обработана ровно один раз. Логирование ошибки - это обработка. Возврат ошибки - это обработка. Не выполняйте обе операции одновременно.

<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> 
<tbody> 
  <tr><td>

  ```golang
  func processUser(id string) error {
    user, err := getUser(id)
    if err != nil {
        log.Printf("Ошибка получения пользователя: %v", err)
        return err // Двойная обработка!
    }
    return nil
  }
  ```
  </td><td>

  ```go
  func processUser(id string) error {
      user, err := getUser(id)
      if err != nil {
          return fmt.Errorf("получить пользователя %s: %w", id, err)
      }
      return nil
  }

  // На верхнем уровне
  func main() {
      if err := processUser("123"); err != nil {
          log.Printf("Ошибка: %v", err)
      }
  }
  ```
  </td></tr> 
</tbody></table>

### Используйте контекст при оборачивании ошибок

Добавляйте контекст к ошибкам, но избегайте избыточных фраз вроде "failed to", "unable to".

<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> <tbody> <tr><td>

```go
if err := saveConfig(cfg); err != nil {
    return fmt.Errorf("failed to save config: %w", err)
}
```
</td><td>

```go
if err := saveConfig(cfg); err != nil {
    return fmt.Errorf("сохранить конфигурацию: %w", err)
}
```
</td></tr> </tbody></table>


### Типы ошибок и их применение

#### Статические ошибки
Используйте `errors.New` для статических строковых ошибок. Экспортируйте их как переменные, если вызывающая сторона должна обрабатывать их специальным образом.

```go
// Пакетные ошибки
var (
    ErrUserNotFound = errors.New("пользователь не найден")
    ErrInvalidToken = errors.New("неверный токен")
)

func GetUser(id string) (*User, error) {
    if id == "" {
        return nil, ErrUserNotFound
    }
    // ...
}
```

#### Динамические ошибки
Используйте `fmt.Errorf` для ошибок с динамическим содержимым или создавайте кастомные типы ошибок, если вызывающая сторона должна их проверять.

<table> <thead><tr><th>Без проверки типа</th><th>С проверкой типа</th></tr></thead> <tbody> <tr><td>

```go
func ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return fmt.Errorf("email %q имеет неверный формат", email)
    }
    return nil
}
```
</td><td>

```go
type ValidationError struct {
    Field   string
    Value   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("поле %s: %s (значение: %q)", 
        e.Field, e.Message, e.Value)
}

func ValidateEmail(email string) error {
    if !strings.Contains(email, "@") {
        return &ValidationError{
            Field:   "email",
            Value:   email,
            Message: "должен содержать @",
        }
    }
    return nil
}
```
</td></tr> </tbody></table>

#### Проверка типов ошибок
Используйте `errors.Is` и `errors.As` для проверки типов ошибок.

```go
func HandleUserRequest(id string) error {
    user, err := getUser(id)
    if err != nil {
        if errors.Is(err, ErrUserNotFound) {
            // Специальная обработка для ненайденного пользователя
            return createDefaultUser(id)
        }
        var valErr *ValidationError
        if errors.As(err, &valErr) {
            // Специальная обработка для ошибок валидации
            return fmt.Errorf("неверный запрос: %s", valErr.Message)
        }
        // Все остальные ошибки
        return fmt.Errorf("обработать запрос: %w", err)
    }
    // ...
}
```

## Использование стандартных библиотек и инструментов разработки
### Стандартные библиотеки
#### Обработка времени (time)
Всегда используйте пакет time для работы со временем. Избегайте самостоятельных расчетов, основанных на некорректных предположениях.

<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> <tbody> <tr><td>

```go
func isActive(now, start, stop int64) bool {
    return start <= now && now < stop
}
```
</td><td>

```go
func isActive(now, start, stop time.Time) bool {
    return (start.Before(now) || start.Equal(now)) && now.Before(stop)
}
```
</td></tr> </tbody></table>

Использование time.Duration
<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> <tbody> <tr><td>

```go
func poll(delay int) {
    for {
        time.Sleep(time.Duration(delay) * time.Millisecond)
    }
}
poll(10) // секунды или миллисекунды?
```
</td><td>

```go
func poll(delay time.Duration) {
    for {
        time.Sleep(delay)
    }
}
poll(10 * time.Second)
```
</td></tr> </tbody></table>

#### Работа с JSON (encoding/json)
Всегда используйте теги структур для сериализации JSON.

<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> <tbody> <tr><td>

```go
type User struct {
    Name string
    Age  int
}
// {"Name": "John", "Age": 30}
```
</td><td>

```go
type User struct {
    Name string `json:"name"`
    Age  int    `json:"age"`
    // Безопасно для переименования полей
}

// {"name": "John", "age": 30}
```
</td></tr> </tbody></table>

#### Конвертация строк (strconv)
Используйте strconv вместо fmt для преобразования примитивных типов.

<table> <thead><tr><th>Плохо</th><th>Хорошо</th></tr></thead> <tbody> <tr><td>

```go
s := fmt.Sprint(123)
```
</td><td>

```go
s := strconv.Itoa(123)
```
</td></tr> <tr><td>

```go
i, _ := strconv.Atoi(fmt.Sprintf("%d", 123))
```
</td><td>

```go
i := 123
s := strconv.Itoa(i)
```
</td></tr> </tbody></table>

#### Работа с файловой системой (os, io)
Всегда проверяйте ошибки и используйте defer для очистки ресурсов.

```go
func readFile(path string) ([]byte, error) {
    f, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("открыть файл: %w", err)
    }
    defer f.Close() // Гарантированное закрытие

    data, err := io.ReadAll(f)
    if err != nil {
        return nil, fmt.Errorf("прочитать файл: %w", err)
    }

    return data, nil
}
```

### Инструменты разработки
#### Форматирование кода
gofmt/goimports
Всегда используйте goimports для автоматического форматирования кода и управления импортами.

Установка
```bash
go install golang.org/x/tools/cmd/goimports@latest
```

Использование
```bash
goimports -local github.com/SlepoyShaman/FileStorage -w .
```

#### Статический анализ
Базовая настройка линтеров

Установка golangci-lint
```bash
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.54.2
```

Запуск
```bash
golangci-lint run
```

#### Тестирование
Табличные тесты
Используйте табличные тесты для покрытия множества сценариев.

```go
func TestParseDuration(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    time.Duration
        wantErr bool
    }{
        {
            name:    "valid duration",
            input:   "1h30m",
            want:    90 * time.Minute,
            wantErr: false,
        },
        {
            name:    "invalid duration",
            input:   "invalid",
            want:    0,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := time.ParseDuration(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseDuration() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

Параллельные тесты
```go
func TestParallel(t *testing.T) {
    tests := []struct {
        name string
        val  int
    }{
        {"test1", 1},
        {"test2", 2},
    }

    for _, tt := range tests {
        tt := tt // Важно: захват переменной
        t.Run(tt.name, func(t *testing.T) {
            t.Parallel()
            // Тестовый код
        })
    }
}
```

## Использование паттернов проектирования для улучшения архитектуры программы
### 1. Функциональные опции (Functional Options)
Суть паттерна: Позволяет создавать гибко настраиваемые объекты без изменения их API и без гигантских конструкторов.

Проблема: При добавлении новых параметров в конструктор приходится менять его сигнатуру, что ломает обратную совместимость.

Решение: Использовать вариадические функции-опции, которые модифицируют объект.

```go
// Базовый объект
type Server struct {
    host    string
    port    int
    timeout time.Duration
}

// Функциональная опция
type Option func(*Server)

func WithTimeout(timeout time.Duration) Option {
    return func(s *Server) { s.timeout = timeout }
}

// Конструктор с опциями
func NewServer(host string, port int, opts ...Option) *Server {
    s := &Server{
        host:    host,
        port:    port,
        timeout: 30 * time.Second, // значение по умолчанию
    }
    
    for _, opt := range opts {
        opt(s) // применяем все опции
    }
    return s
}

// Использование
server := NewServer("localhost", 8080, 
    WithTimeout(60*time.Second),
)
```

Преимущества:
- Не ломает обратную совместимость
- Читаемый код создания объектов
- Легко добавлять новые параметры

### 2. Dependency Injection (Внедрение зависимостей)
Суть паттерна: Передавать зависимости извне, а не создавать их внутри класса.

Проблема: Жесткая связь между компонентами, сложное тестирование.

Решение: Зависимости передаются через конструктор.

```go
// Бизнес-логика
type UserService struct {
    repo     UserRepository  // зависимость
    notifier Notifier        // зависимость  
}

// Зависимости передаются явно
func NewUserService(repo UserRepository, notifier Notifier) *UserService {
    return &UserService{
        repo:     repo,
        notifier: notifier,
    }
}

func (s *UserService) Register(user *User) error {
    // Используем переданные зависимости
    if err := s.repo.Save(user); err != nil {
        return err
    }
    return s.notifier.SendWelcome(user.Email)
}

// Использование
repo := NewUserRepository(db)
notifier := NewEmailNotifier()
service := NewUserService(repo, notifier) // зависимости внедряются
```

Преимущества:
- Легкое тестирование (можно передавать моки)
- Гибкая замена реализаций
- Четкие границы ответственности

### 3. Repository Pattern
Суть паттерна: Абстрагирует доступ к данным от бизнес-логики.

Проблема: Бизнес-логика завязана на конкретную базу данных.

Решение: Создать интерфейс для операций с данными.

```go
// Доменная сущность
type User struct {
    ID    string
    Name  string
    Email string
}

// Абстракция для работы с данными
type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Конкретная реализация для MySQL
type MySQLUserRepository struct {
    db *sql.DB
}

func (r *MySQLUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    // Реализация для MySQL
    var user User
    err := r.db.QueryRowContext(ctx, "SELECT ...", id).Scan(&user.ID, &user.Name, &user.Email)
    return &user, err
}

// Реализация для тестов
type InMemoryUserRepository struct {
    users map[string]*User
}

func (r *InMemoryUserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    user, exists := r.users[id]
    if !exists {
        return nil, errors.New("user not found")
    }
    return user, nil
}

// Бизнес-логика работает с интерфейсом
type UserService struct {
    repo UserRepository // абстракция, а не конкретная реализация
}
```
Преимущества:
- Бизнес-логика не зависит от БД
- Легко менять хранилище данных

## Управление версиями кода с использованием систем контроля версий, таких как Git.

### Основные ветки:

main - стабильная версия приложения

MAI-N-* - разработка новых функций (MAI-N - номер задачи в проекте)

MAI-N-hotfix-* - срочные исправления 

### Создание feature ветки
```bash
git checkout -b MAI-1-user-dashboard main
```

### Создание hotfix ветки
```bash
git checkout -b MAI-1-hotfix-critical-security main
```

### Получить последние изменения
```bash
git fetch origin
git rebase origin/main
```

### Просмотр состояния
```bash
git status
git log --oneline -10
```

### Подготовка коммита
```bash
git add -p  # интерактивное добавление изменений
git commit -m "MAI-N описание изменения"
```

### Отправка изменений
```bash
git push origin MAI-N-branch-name
```

### Разрешение конфликтов

Во время rebase
```bash
git rebase origin/main
# ... конфликт ...
git status  # показать конфликтующие файлы
# Редактируем файлы, разрешаем конфликты
git add resolved-file.go
git rebase --continue
```

### Пулл-реквесты
Структура PR
Название: MAI-99 add JWT token support

Описание:

**Что сделано**
- Добавлена JWT аутентификация
- Реализован middleware для проверки токенов

**Чеклист**
- [ ] Код покрыт тестами
- [ ] Добавлена документация
- [ ] Обновлены зависимости

### Изменение истории коммитов
```bash
git rebase -i HEAD~3
```

### Редактирование последних 3 коммитов:
```bash
pick a1b2c3 feat: add feature A
reword d4e5f6 fix: typo in documentation
squash g7h8i9 docs: update examples
```

**Команды:**
- pick - использовать коммит
- reword - использовать, изменить сообщение
- squash - объединить с предыдущим
- fixup - объединить, отбросить сообщение


### **Басангов Александр Лабораторная работа 3**

## Использование комментариев для пояснения сложных участков кода

### Общие принципы комментирования

### Когда комментировать
- **Сложная бизнес-логика**: Алгоритмы с неочевидной реализацией
- **Нетрадиционные решения**: Когда код отклоняется от стандартных подходов
- **Внешние зависимости**: Интеграции со сторонними системами
- **Временные решения**: Код, который требует последующего рефакторинга
- **Неочевидные ограничения**: Ограничения, не следующие из сигнатуры функции

### Когда не комментировать
- **Простые операции**: `i++` или `user.Name = "John"`
- **Очевидная логика**: Код, который легко читается и понимается
- **Хорошие имена**: Когда имена переменных и функций достаточно описательны

### Типы комментариев

### Блочные комментарии для сложной логики
```go
// calculateOptimalRoute вычисляет оптимальный маршрут с учетом множества ограничений
// Алгоритм основан на модифицированном алгоритме Дейкстры с эвристиками:
// 1. Приоритет дорог с меньшим трафиком
// 2. Избегание платных участков при наличии бесплатных альтернатив
// 3. Учет текущих дорожных событий (аварии, ремонты)
//
// Сложность: O((V + E) log V), где V - вершины, E - ребра графа
func calculateOptimalRoute(start, end coordinates, constraints RouteConstraints) (*Route, error) {
    graph := initializeRouteGraph()
    applyTollHeuristics(graph, constraints.AvoidTolls)
    applyTrafficIncidents(graph, getCurrentIncidents())
    return findShortestPath(graph, start, end)
}
```

### Inline комментарии для неочевидных решений
```go
func processFinancialTransaction(tx *Transaction) error {
    // Используем UTC чтобы избежать проблем с летним временем
    now := time.Now().UTC()
    
    // Округление до 4 знаков для финансовых расчетов
    // Требование регулятора - точность до 0.0001
    amount := roundToFourDecimals(tx.Amount)

    if tx.IsExternal {
        // Комиссия 1.5%, но не менее 50 рублей
        fee := calculateFee(amount, 0.015, 50.0)
        amount -= fee
    }
    
    // Используем блокировку для избежания гонок при одновременных операциях с одним счетом
    lockKey := fmt.Sprintf("account_%s", tx.AccountID)
    mutex := getDistributedLock(lockKey)
    defer mutex.Unlock()
    
    return executeTransaction(tx.AccountID, amount)
}
```


### API интеграции
```go
// processPayment обрабатывает платеж через внешнюю платежную систему
// Особенности интеграции:
// - Таймаут 30 секунд на операцию
// - Retry с экспоненциальной backoff стратегией
// - Обязательная idempotency key для избежания дублирования
// - Логирование всех запросов для аудита
//
// Документация API: https://payments.example.com/docs/v2#process-payment
func processPayment(payment PaymentRequest) (*PaymentResponse, error) {
    // Генерация idempotency key для гарантии идемпотентности
    // При повторных вызовах с тем же ключом возвращается тот же результат
    idempotencyKey := generateIdempotencyKey(payment.OrderID)
    
    client := &http.Client{
        Timeout: 30 * time.Second,
    }
    
    // Retry стратегия: 3 попытки с задержкой 1s, 2s, 4s
    var lastErr error
    for i := 0; i < 3; i++ {
        if i > 0 {
            time.Sleep(time.Duration(math.Pow(2, float64(i-1))) * time.Second)
        }
        
        resp, err := sendPaymentRequest(client, payment, idempotencyKey)
        if err == nil {
            return resp, nil
        }
        lastErr = err
        
        // Не повторяем для клиентских ошибок 4xx
        if isClientError(err) {
            break
        }
    }
    
    return nil, fmt.Errorf("payment processing failed after retries: %w", lastErr)
}
```

### Сложная обработка ошибок
```go
// validateAndProcess комплексная валидация и обработка заказа
// Возвращает детализированные ошибки для разных сценариев:
// - InsufficientFundsError: недостаточно средств
// - InventoryError: товара нет в наличии  
// - ValidationError: ошибки валидации данных
// - SystemError: внутренние ошибки системы
func validateAndProcess(order Order) error {
    // Валидация базовых полей заказа
    if err := validateOrderStructure(order); err != nil {
        return ValidationError{Field: "structure", Reason: err.Error()}
    }
    
    // Проверка доступности товаров на складе
    // Используем пессимистические блокировки для избежания race condition
    available, err := checkInventoryWithLock(order.Items)
    if err != nil {
        return SystemError{Operation: "inventory_check", Cause: err}
    }
    if !available {
        return InventoryError{Items: getUnavailableItems(order.Items)}
    }
    
    // Проверка баланса пользователя
    // Учитываем резервирование средств других pending заказов
    balance, err := getUserAvailableBalance(order.UserID)
    if err != nil {
        return SystemError{Operation: "balance_check", Cause: err}
    }
    
    total := calculateOrderTotal(order)
    if balance < total {
        shortfall := total - balance
        return InsufficientFundsError{
            CurrentBalance: balance,
            Required:       total,
            Shortfall:      shortfall,
        }
    }
    
    // Основная обработка заказа
    return processOrderTransaction(order)
}
```

## Проверка типов данных и корректности ввода данных

### Уровни валидации
- **Входные данные**: Валидация на границах системы (API, CLI)
- **Бизнес-логика**: Проверки в сервисном слое
- **База данных**: Ограничения на уровне СУБД

### Валидируем входные данные
```go
// Плохо: доверяем входным данным
func ProcessUser(input map[string]interface{}) error {
    name := input["name"].(string) // может panic
    age := input["age"].(int)      // может panic
    // обработка...
}

// Хорошо: проверяем и валидируем
func ProcessUser(input map[string]interface{}) error {
    if err := validateUserInput(input); err != nil {
        return fmt.Errorf("invalid input: %w", err)
    }
    
    name, _ := input["name"].(string)
    age, _ := input["age"].(int)
    // обработка...
}
```

### Базовая валидация структур
```go
type UserRegistrationRequest struct {
    Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8,containsany=!@#$%"`
    Age      int    `json:"age" validate:"required,min=18,max=120"`
    Phone    string `json:"phone" validate:"omitempty,e164"`
}

// Validate выполняет валидацию структуры
func (r *UserRegistrationRequest) Validate() error {
    if strings.TrimSpace(r.Username) == "" {
        return errors.New("username is required")
    }
    
    if len(r.Username) < 3 || len(r.Username) > 50 {
        return errors.New("username must be between 3 and 50 characters")
    }
    
    if !isAlphanumeric(r.Username) {
        return errors.New("username must contain only letters and numbers")
    }
    
    if !isValidEmail(r.Email) {
        return errors.New("invalid email format")
    }
    
    if len(r.Password) < 8 {
        return errors.New("password must be at least 8 characters")
    }
    
    if !containsSpecialChar(r.Password) {
        return errors.New("password must contain at least one special character")
    }
    
    if r.Age < 18 || r.Age > 120 {
        return errors.New("age must be between 18 and 120")
    }
    
    if r.Phone != "" && !isValidPhone(r.Phone) {
        return errors.New("invalid phone format")
    }
    
    return nil
}
```

### Кастомные валидаторы
```go
// CurrencyValidator проверяет корректность валюты
type CurrencyValidator struct{}

func (v CurrencyValidator) Validate(value interface{}) error {
    currency, ok := value.(string)
    if !ok {
        return errors.New("currency must be a string")
    }
    
    supportedCurrencies := map[string]bool{
        "USD": true, "EUR": true, "GBP": true, "JPY": true,
        "CAD": true, "AUD": true, "CHF": true, "CNY": true,
    }
    
    if !supportedCurrencies[strings.ToUpper(currency)] {
        return fmt.Errorf("unsupported currency: %s", currency)
    }
    
    return nil
}

// DateRangeValidator проверяет корректность диапазона дат
type DateRangeValidator struct{}

func (v DateRangeValidator) Validate(start, end time.Time) error {
    if start.IsZero() || end.IsZero() {
        return errors.New("both start and end dates are required")
    }
    
    if end.Before(start) {
        return errors.New("end date cannot be before start date")
    }
    
    if end.Sub(start) > 365*24*time.Hour {
        return errors.New("date range cannot exceed 1 year")
    }
    
    return nil
}
```

### Уровень HTTP обработчика
```go
type UserHandler struct {
    service UserService
}

// CreateUser обрабатывает создание пользователя
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req UserRegistrationRequest
    
    // Валидация Content-Type
    if r.Header.Get("Content-Type") != "application/json" {
        http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
        return
    }
    
    // Парсинг и валидация JSON
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
        return
    }
    
    // Валидация бизнес-логики
    if err := req.Validate(); err != nil {
        http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
        return
    }
    
    // Дополнительные проверки
    if err := h.validateRateLimit(r); err != nil {
        http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
        return
    }
    
    // Обработка запроса
    user, err := h.service.RegisterUser(r.Context(), req)
    if err != nil {
        handleServiceError(w, err)
        return
    }
    
    respondJSON(w, user, http.StatusCreated)
}
```


### Безопасное приведение типов
```go

// ToInt безопасно преобразует interface{} в int
func ToInt(value interface{}) (int, error) {
    switch v := value.(type) {
    case int:
        return v, nil
    case int64:
        if v > math.MaxInt || v < math.MinInt {
            return 0, fmt.Errorf("int64 value %d out of int range", v)
        }
        return int(v), nil
    case float64:
        if v != math.Trunc(v) {
            return 0, errors.New("float value has fractional part")
        }
        if v > math.MaxInt || v < math.MinInt {
            return 0, fmt.Errorf("float value %f out of int range", v)
        }
        return int(v), nil
    case string:
        return strconv.Atoi(v)
    default:
        return 0, fmt.Errorf("unsupported type: %T", value)
    }
}

// ToString безопасно преобразует interface{} в string
func ToString(value interface{}) (string, error) {
    switch v := value.(type) {
    case string:
        return v, nil
    case int, int64, float64, bool:
        return fmt.Sprintf("%v", v), nil
    default:
        return "", fmt.Errorf("cannot convert type %T to string", value)
    }
}
```

### Валидация JSON схем
```go
// JSONSchemaValidator проверяет JSON по схеме
type JSONSchemaValidator struct {
    schemas map[string]jsonschema.Schema
}

func NewJSONSchemaValidator() *JSONSchemaValidator {
    return &JSONSchemaValidator{
        schemas: loadSchemas(),
    }
}

// Validate проверяет JSON данные по указанной схеме
func (v *JSONSchemaValidator) Validate(schemaName string, data []byte) error {
    schema, exists := v.schemas[schemaName]
    if !exists {
        return fmt.Errorf("schema %s not found", schemaName)
    }
    
    var jsonData interface{}
    if err := json.Unmarshal(data, &jsonData); err != nil {
        return fmt.Errorf("invalid JSON: %w", err)
    }
    
    if err := schema.Validate(jsonData); err != nil {
        return fmt.Errorf("schema validation failed: %w", err)
    }
    
    return nil
}

// Пример использования
func validateUserJSON(data []byte) error {
    validator := NewJSONSchemaValidator()
    
    if err := validator.Validate("user_registration", data); err != nil {
        return fmt.Errorf("user data validation failed: %w", err)
    }
    
    return nil
}
```

## Работа с параллельными процессами и многопоточностью


### Горутины 
```go
// Запуск простой горутины
func main() {
    // Плохо: запуск горутины без контроля завершения
    go processData(data)
    
    // Хорошо: с механизмом ожидания завершения
    var wg sync.WaitGroup
    wg.Add(1)
    
    go func() {
        defer wg.Done()
        processData(data)
    }()
    
    wg.Wait()
}

// Запуск множества горутин
func processBatch(items []Item) error {
    var wg sync.WaitGroup
    errCh := make(chan error, len(items))
    
    for _, item := range items {
        wg.Add(1)
        
        go func(item Item) {
            defer wg.Done()
            
            if err := processItem(item); err != nil {
                errCh <- fmt.Errorf("processing item %v: %w", item.ID, err)
            }
        }(item) // Важно: передавать параметр в горутину
    }
    
    wg.Wait()
    close(errCh)
    
    // Сбор ошибок
    var errors []error
    for err := range errCh {
        errors = append(errors, err)
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("batch processing failed: %v", errors)
    }
    
    return nil
}
```

### Мьютексы
```go
// Thread-safe кэш с использованием RWMutex
type SafeCache struct {
    mu    sync.RWMutex
    data  map[string]interface{}
    stats CacheStats
}

func NewSafeCache() *SafeCache {
    return &SafeCache{
        data:  make(map[string]interface{}),
        stats: CacheStats{},
    }
}

// Get безопасное чтение с блокировкой чтения
func (c *SafeCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    value, exists := c.data[key]
    if exists {
        c.stats.Hits++
    } else {
        c.stats.Misses++
    }
    
    return value, exists
}

// Set безопасная запись с блокировкой записи
func (c *SafeCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    c.data[key] = value
    c.stats.Writes++
}

// GetAll безопасное получение всех данных
func (c *SafeCache) GetAll() map[string]interface{} {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    // Возвращаем копию для избежания гонок данных
    result := make(map[string]interface{})
    for k, v := range c.data {
        result[k] = v
    }
    
    return result
}
```

### Worker Pool
```go
// WorkerPool реализует пул воркеров для обработки задач
type WorkerPool struct {
    workers   int
    taskQueue chan Task
    resultCh  chan Result
    done      chan struct{}
    wg        sync.WaitGroup
}

type Task struct {
    ID   int
    Data interface{}
}

type Result struct {
    TaskID int
    Data   interface{}
    Error  error
}

func NewWorkerPool(workers int) *WorkerPool {
    return &WorkerPool{
        workers:   workers,
        taskQueue: make(chan Task, 100),
        resultCh:  make(chan Result, 100),
        done:      make(chan struct{}),
    }
}

// Start запускает пул воркеров
func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        wp.wg.Add(1)
        
        go wp.worker(i)
    }
    
    // Горутина для закрытия resultCh после завершения всех воркеров
    go func() {
        wp.wg.Wait()
        close(wp.resultCh)
    }()
}

func (wp *WorkerPool) worker(id int) {
    defer wp.wg.Done()
    
    for {
        select {
        case task, ok := <-wp.taskQueue:
            if !ok {
                return // Канал закрыт, завершаем работу
            }
            
            // Обработка задачи
            result := wp.processTask(task)
            wp.resultCh <- result
            
        case <-wp.done:
            return // Получен сигнал завершения
        }
    }
}

func (wp *WorkerPool) processTask(task Task) Result {
    // Имитация обработки
    time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
    
    // В 10% случаев возвращаем ошибку
    if rand.Float32() < 0.1 {
        return Result{
            TaskID: task.ID,
            Error:  fmt.Errorf("processing failed for task %d", task.ID),
        }
    }
    
    return Result{
        TaskID: task.ID,
        Data:   fmt.Sprintf("processed: %v", task.Data),
    }
}

// Submit добавляет задачу в очередь
func (wp *WorkerPool) Submit(task Task) {
    wp.taskQueue <- task
}

// Results возвращает канал с результатами
func (wp *WorkerPool) Results() <-chan Result {
    return wp.resultCh
}

// Stop останавливает пул воркеров
func (wp *WorkerPool) Stop() {
    close(wp.done)
    close(wp.taskQueue)
}

// Пример использования
func main() {
    pool := NewWorkerPool(5)
    pool.Start()
    
    // Отправка задач
    for i := 0; i < 20; i++ {
        pool.Submit(Task{
            ID:   i,
            Data: fmt.Sprintf("task_data_%d", i),
        })
    }
    
    // Сбор результатов
    go func() {
        for result := range pool.Results() {
            if result.Error != nil {
                log.Printf("Task %d failed: %v", result.TaskID, result.Error)
            } else {
                log.Printf("Task %d completed: %v", result.TaskID, result.Data)
            }
        }
    }()
    
    // Даем время на обработку
    time.Sleep(2 * time.Second)
    pool.Stop()
}
```

## Рефакторинг кода для улучшения его читаемости, эффективности и поддерживаемости

### Когда проводить рефакторинг
- **Перед добавлением новой функциональности** - подготовка кода к изменениям
- **После code review** - исправление замечаний ревьюеров
- **При обнаружении дублирования** - устранение копипасты
- **Когда код сложно понять** - улучшение читаемости
- **При падении производительности** - оптимизация критических участков

### Выделение методов и функций
```go
// Плохо: большой метод со смешанной ответственностью
func ProcessUserOrder(userID string, orderData map[string]interface{}) error {
    // Валидация пользователя
    user, err := db.GetUser(userID)
    if err != nil {
        return err
    }
    if user == nil {
        return errors.New("user not found")
    }
    if !user.IsActive {
        return errors.New("user is not active")
    }
    
    // Валидация заказа
    items, ok := orderData["items"].([]interface{})
    if !ok {
        return errors.New("invalid items format")
    }
    if len(items) == 0 {
        return errors.New("no items in order")
    }
    
    // Расчет стоимости
    var total float64
    for _, item := range items {
        itemMap, ok := item.(map[string]interface{})
        if !ok {
            return errors.New("invalid item format")
        }
        price, ok := itemMap["price"].(float64)
        if !ok {
            return errors.New("invalid price format")
        }
        quantity, ok := itemMap["quantity"].(float64)
        if !ok {
            return errors.New("invalid quantity format")
        }
        total += price * quantity
    }
    
    // Применение скидки
    if user.IsPremium {
        total *= 0.9 // 10% скидка
    }
    
    // Сохранение заказа
    order := Order{
        UserID: userID,
        Total:  total,
        Items:  convertItems(items),
    }
    
    return db.SaveOrder(order)
}

// Хорошо: разделение на маленькие методы с четкой ответственностью
func ProcessUserOrder(userID string, orderData OrderData) error {
    user, err := validateAndGetUser(userID)
    if err != nil {
        return fmt.Errorf("user validation failed: %w", err)
    }
    
    items, err := validateOrderItems(orderData.Items)
    if err != nil {
        return fmt.Errorf("order validation failed: %w", err)
    }
    
    total := calculateOrderTotal(items, user)
    order := createOrder(userID, items, total)
    
    return saveOrder(order)
}

// Маленькие, сфокусированные методы
func validateAndGetUser(userID string) (*User, error) {
    user, err := db.GetUser(userID)
    if err != nil {
        return nil, err
    }
    if user == nil {
        return nil, ErrUserNotFound
    }
    if !user.IsActive {
        return nil, ErrUserInactive
    }
    return user, nil
}

func validateOrderItems(rawItems []OrderItem) ([]ValidatedItem, error) {
    if len(rawItems) == 0 {
        return nil, ErrEmptyOrder
    }
    
    var validatedItems []ValidatedItem
    for i, item := range rawItems {
        validated, err := validateItem(item, i)
        if err != nil {
            return nil, err
        }
        validatedItems = append(validatedItems, validated)
    }
    
    return validatedItems, nil
}
```

### Улучшение имен переменных и функций
```go
// Плохо: неясные имена
func proc(u *User, d map[string]interface{}) error {
    n := d["n"].(string)
    a := d["a"].(int)
    // что такое u, d, n, a?
}

// Хорошо: описательные имена
func ProcessUserRegistration(user *User, registrationData map[string]interface{}) error {
    userName := registrationData["name"].(string)
    userAge := registrationData["age"].(int)
    // ясно что происходит
}
```

### Упрощение условных выражений
```go
// Плохо: сложные вложенные условия
func CanUserAccessResource(user *User, resource *Resource) bool {
    if user != nil {
        if user.IsAdmin {
            return true
        } else {
            if resource != nil {
                if resource.OwnerID == user.ID {
                    return true
                } else {
                    for _, perm := range user.Permissions {
                        if perm.ResourceID == resource.ID && perm.CanRead {
                            return true
                        }
                    }
                }
            }
        }
    }
    return false
}

// Хорошо: упрощенные условия с guard clauses
func CanUserAccessResource(user *User, resource *Resource) bool {
    if user == nil || resource == nil {
        return false
    }
    
    if user.IsAdmin {
        return true
    }
    
    if resource.OwnerID == user.ID {
        return true
    }
    
    return user.HasReadPermission(resource.ID)
}
```

### Оптимизация алгоритмов
```go
// Плохо: неэффективный поиск O(n²)
func FindCommonItems(list1, list2 []string) []string {
    var common []string
    for _, item1 := range list1 {
        for _, item2 := range list2 {
            if item1 == item2 {
                common = append(common, item1)
                break
            }
        }
    }
    return common
}

// Хорошо: использование map для O(n) поиска
func FindCommonItems(list1, list2 []string) []string {
    if len(list1) == 0 || len(list2) == 0 {
        return nil
    }
    
    // Создаем set из второго списка
    set := make(map[string]bool, len(list2))
    for _, item := range list2 {
        set[item] = true
    }
    
    // Ищем пересечения
    var common []string
    for _, item := range list1 {
        if set[item] {
            common = append(common, item)
        }
    }
    
    return common
}
```

### Замена кодов ошибок на типизированные ошибки
```go
// Плохо: магические числа и строки
func ProcessPayment(amount float64) error {
    if amount <= 0 {
        return errors.New("invalid amount")
    }
    
    if amount > 10000 {
        return errors.New("amount too large")
    }
    
    // обработка платежа...
}

// Хорошо: типизированные ошибки
var (
    ErrInvalidAmount   = errors.New("invalid amount")
    ErrAmountTooLarge  = errors.New("amount exceeds limit")
    ErrInsufficientFunds = errors.New("insufficient funds")
)

type PaymentError struct {
    Operation string
    Amount    float64
    Reason    error
}

func (e PaymentError) Error() string {
    return fmt.Sprintf("payment operation %s failed for amount %.2f: %v", 
        e.Operation, e.Amount, e.Reason)
}

func (e PaymentError) Unwrap() error {
    return e.Reason
}

func ProcessPayment(amount float64) error {
    if amount <= 0 {
        return PaymentError{
            Operation: "validation",
            Amount:    amount,
            Reason:    ErrInvalidAmount,
        }
    }
    
    if amount > 10000 {
        return PaymentError{
            Operation: "validation", 
            Amount:    amount,
            Reason:    ErrAmountTooLarge,
        }
    }
    
    // обработка платежа...
}
```