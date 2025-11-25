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
