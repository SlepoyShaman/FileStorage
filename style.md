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
