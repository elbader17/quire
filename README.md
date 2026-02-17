# Quire

Una librería de Go que proporciona una interfaz tipo base de datos para Google Sheets. Convierte tus hojas de cálculo en una base de datos documental con operaciones CRUD, consultas con filtros y mapeo de structs.

[![Go Report Card](https://goreportcard.com/badge/github.com/yourusername/quire)](https://goreportcard.com/report/github.com/yourusername/quire)
[![GoDoc](https://godoc.org/github.com/yourusername/quire?status.svg)](https://godoc.org/github.com/yourusername/quire)

## Características

- **API tipo ORM**: Operaciones familiares de base de datos (CRUD) sobre Google Sheets
- **Mapeo type-safe**: Convierte structs de Go a filas de sheets usando tags
- **Query builder**: Construye consultas fluidas con filtros, límites y ordenamiento
- **Autenticación segura**: Usa Service Accounts para autenticación server-to-server
- **Sin esquema**: No requiere definir tablas ni columnas previamente
- **Concurrente**: Soporte completo para operaciones concurrentes con `context.Context`

## Tabla de Contenidos

- [Instalación](#instalación)
- [Configuración de Google Cloud](#configuración-de-google-cloud)
- [Guía de Inicio Rápido](#guía-de-inicio-rápido)
- [Conceptos Fundamentales](#conceptos-fundamentales)
- [API Completa](#api-completa)
  - [Configuración](#configuración)
  - [Conexión](#conexión)
  - [Inserción de Datos](#inserción-de-datos)
  - [Consultas](#consultas)
  - [Filtros](#filtros)
  - [Mapeo de Structs](#mapeo-de-structs)
- [Ejemplos Avanzados](#ejemplos-avanzados)
- [Mejores Prácticas](#mejores-prácticas)
- [Limitaciones](#limitaciones)
- [Solución de Problemas](#solución-de-problemas)

## Instalación

```bash
go get github.com/yourusername/quire
```

Requiere Go 1.21 o superior.

## Configuración de Google Cloud

Antes de usar Quire, necesitas configurar el acceso a Google Sheets API:

### 1. Crear un Proyecto en Google Cloud Console

1. Ve a [Google Cloud Console](https://console.cloud.google.com/)
2. Crea un nuevo proyecto o selecciona uno existente
3. Habilita la **Google Sheets API**:
   - Navega a "APIs & Services" > "Library"
   - Busca "Google Sheets API"
   - Haz clic en "Enable"

### 2. Crear una Service Account

1. Ve a "IAM & Admin" > "Service Accounts"
2. Haz clic en "Create Service Account"
3. Asigna un nombre y descripción
4. En "Grant this service account access to project", selecciona "Editor" o rol personalizado con permisos de Sheets
5. Crea la cuenta

### 3. Generar Claves

1. En la lista de Service Accounts, haz clic en la cuenta creada
2. Ve a la pestaña "Keys"
3. Haz clic en "Add Key" > "Create new key"
4. Selecciona formato JSON
5. Descarga el archivo (ej: `service-account.json`)

**IMPORTANTE**: Guarda este archivo de forma segura. No lo commitees en tu repositorio.

### 4. Compartir tu Spreadsheet

1. Abre tu Google Sheet
2. Haz clic en "Share" (Compartir)
3. Agrega el email de la Service Account (está en el JSON, campo `client_email`)
4. Asigna permisos de "Editor"

### 5. Obtener el Spreadsheet ID

El ID está en la URL de tu Google Sheet:

```
https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
                              ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
                              Este es tu Spreadsheet ID
```

## Guía de Inicio Rápido

### Ejemplo Mínimo

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/yourusername/quire/pkg/quire"
)

type User struct {
    ID    int    `quire:"ID"`
    Name  string `quire:"Name"`
    Email string `quire:"Email"`
    Age   int    `quire:"Age"`
}

func main() {
    ctx := context.Background()
    
    // Leer credenciales
    credentials, err := os.ReadFile("service-account.json")
    if err != nil {
        log.Fatal(err)
    }
    
    // Crear conexión
    db, err := quire.New(quire.Config{
        SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms",
        Credentials:   credentials,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Insertar datos
    users := db.Table("Users")
    err = users.Insert(ctx, []User{
        {ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30},
        {ID: 2, Name: "Bob", Email: "bob@example.com", Age: 25},
    })
    if err != nil {
        log.Fatal(err)
    }
    
    // Consultar datos
    var results []User
    err = users.Query().
        Where("Age", ">=", 25).
        Limit(10).
        Get(ctx, &results)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Encontrados %d usuarios", len(results))
}
```

## Conceptos Fundamentales

### Estructura de la Hoja

Quire espera que tu Google Sheet tenga la siguiente estructura:

| ID | Name  | Email             | Age |
|----|-------|-------------------|-----|
| 1  | Alice | alice@example.com | 30  |
| 2  | Bob   | bob@example.com   | 25  |

- **Primera fila**: Headers que coincidan con los tags `quire` de tu struct
- **Filas siguientes**: Datos
- Cada sheet (pestaña) representa una "tabla"

### Mapeo de Tipos

Quire convierte automáticamente entre tipos de Go y valores de Sheets:

| Tipo Go | Ejemplo en Sheet | Conversión |
|---------|------------------|------------|
| `string` | "Alice" | Directa |
| `int` | 42 ó "42" | Parseo desde string o float |
| `float64` | 3.14 ó "3.14" | Parseo desde string |
| `bool` | "true", "TRUE", "1" | Parseo case-insensitive |
| `uint` | 100 ó "100" | Parseo con validación |

## API Completa

### Configuración

```go
type Config struct {
    // SpreadsheetID es el ID de tu Google Sheet (obligatorio)
    // Se encuentra en la URL: .../d/{SPREADSHEET_ID}/...
    SpreadsheetID string
    
    // Credentials es el contenido del archivo JSON de la Service Account (obligatorio)
    // Usa os.ReadFile() para cargar el archivo
    Credentials []byte
}
```

### Conexión

```go
// Crear una nueva instancia de base de datos
db, err := quire.New(quire.Config{
    SpreadsheetID: "tu-spreadsheet-id",
    Credentials:   credentials,
})
if err != nil {
    // Manejar error (credenciales inválidas, formato incorrecto, etc.)
}
defer db.Close()

// Obtener referencia a una tabla (sheet)
users := db.Table("Users")  // "Users" es el nombre de la pestaña
products := db.Table("Products")
```

### Inserción de Datos

```go
// Insert acepta un slice de structs
users := []User{
    {ID: 1, Name: "Alice", Email: "alice@example.com"},
    {ID: 2, Name: "Bob", Email: "bob@example.com"},
}

err := db.Table("Users").Insert(ctx, users)
if err != nil {
    log.Fatal(err)
}
```

**Notas:**
- Los datos se agregan al final de la hoja (append)
- No se verifican duplicados automáticamente
- Los campos con tag `quire:"-"` se ignoran

### Consultas

#### Query Básico

```go
var users []User
err := db.Table("Users").Query().Get(ctx, &users)
```

#### Con Límite

```go
var users []User
err := db.Table("Users").Query().
    Limit(10).
    Get(ctx, &users)
```

#### Con Ordenamiento (placeholder)

```go
var users []User
err := db.Table("Users").Query().
    OrderBy("Age", true).  // true = descendente
    Get(ctx, &users)
```

**Nota:** El ordenamiento actualmente es un placeholder y no ordena los resultados.

### Filtros

#### Operadores Soportados

| Operador | Descripción | Ejemplo |
|----------|-------------|---------|
| `=`, `==` | Igualdad | `Where("Status", "=", "active")` |
| `!=` | Desigualdad | `Where("Status", "!=", "deleted")` |
| `>` | Mayor que | `Where("Age", ">", 18)` |
| `>=` | Mayor o igual | `Where("Score", ">=", 90)` |
| `<` | Menor que | `Where("Price", "<", 100)` |
| `<=` | Menor o igual | `Where("Stock", "<=", 10)` |
| `contains`, `like` | Contiene substring (case-insensitive) | `Where("Name", "contains", "john")` |

#### Múltiples Filtros (AND)

```go
var users []User
err := db.Table("Users").Query().
    Where("Age", ">=", 18).
    Where("Status", "=", "active").
    Where("Country", "=", "US").
    Get(ctx, &users)
// WHERE Age >= 18 AND Status = 'active' AND Country = 'US'
```

#### Filtros con Strings

```go
// Búsqueda case-insensitive
var users []User
err := db.Table("Users").Query().
    Where("Name", "contains", "alice").
    Get(ctx, &users)
// Coincide con "Alice", "ALICE", "alice smith", etc.
```

### Mapeo de Structs

#### Tags Básicos

```go
type Product struct {
    ID          int     `quire:"ID"`          // Mapea a columna "ID"
    Name        string  `quire:"Name"`        // Mapea a columna "Name"
    Price       float64 `quire:"Price"`       // Mapea a columna "Price"
    Description string  `quire:"Description"` // Mapea a columna "Description"
}
```

#### Ignorar Campos

```go
type User struct {
    ID        int       `quire:"ID"`
    Name      string    `quire:"Name"`
    Password  string    `quire:"-"`  // Este campo se ignora
    CreatedAt time.Time `quire:"-"`  // Este campo también se ignora
}
```

#### Mapeo por Nombre de Campo

Si no especificas tag, se usa el nombre del campo:

```go
type User struct {
    ID   int    // Mapea a columna "ID"
    Name string // Mapea a columna "Name"
}
```

## Ejemplos Avanzados

### CRUD Completo

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    
    "github.com/yourusername/quire/pkg/quire"
)

type Task struct {
    ID          int    `quire:"ID"`
    Title       string `quire:"Title"`
    Description string `quire:"Description"`
    Status      string `quire:"Status"`
    Priority    int    `quire:"Priority"`
}

func main() {
    ctx := context.Background()
    
    credentials, _ := os.ReadFile("service-account.json")
    db, _ := quire.New(quire.Config{
        SpreadsheetID: "tu-spreadsheet-id",
        Credentials:   credentials,
    })
    defer db.Close()
    
    tasks := db.Table("Tasks")
    
    // CREATE
    newTasks := []Task{
        {ID: 1, Title: "Implementar login", Status: "pending", Priority: 1},
        {ID: 2, Title: "Crear tests", Status: "pending", Priority: 2},
        {ID: 3, Title: "Documentar API", Status: "done", Priority: 3},
    }
    
    if err := tasks.Insert(ctx, newTasks); err != nil {
        log.Fatal(err)
    }
    fmt.Println("Tareas creadas")
    
    // READ - Todas las tareas pendientes de alta prioridad
    var pendingTasks []Task
    err := tasks.Query().
        Where("Status", "=", "pending").
        Where("Priority", "<=", 2).
        Get(ctx, &pendingTasks)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Tareas pendientes de alta prioridad: %d\n", len(pendingTasks))
    for _, task := range pendingTasks {
        fmt.Printf("  - %s (Prioridad: %d)\n", task.Title, task.Priority)
    }
    
    // READ - Tareas completadas
    var doneTasks []Task
    err = tasks.Query().
        Where("Status", "=", "done").
        Get(ctx, &doneTasks)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Tareas completadas: %d\n", len(doneTasks))
}
```

### Sistema de Inventario

```go
type Product struct {
    SKU       string  `quire:"SKU"`
    Name      string  `quire:"Name"`
    Category  string  `quire:"Category"`
    Price     float64 `quire:"Price"`
    Stock     int     `quire:"Stock"`
}

func getLowStockProducts(ctx context.Context, db *quire.DB) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Stock", "<", 10).
        OrderBy("Stock", false).  // Ordenar por stock ascendente
        Get(ctx, &products)
    return products, err
}

func getProductsByCategory(ctx context.Context, db *quire.DB, category string) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Category", "=", category).
        Where("Stock", ">", 0).  // Solo productos disponibles
        Get(ctx, &products)
    return products, err
}

func searchProducts(ctx context.Context, db *quire.DB, query string) ([]Product, error) {
    var products []Product
    err := db.Table("Inventory").Query().
        Where("Name", "contains", query).
        Limit(20).
        Get(ctx, &products)
    return products, err
}
```

### Gestión de Usuarios

```go
type User struct {
    ID        int    `quire:"ID"`
    Username  string `quire:"Username"`
    Email     string `quire:"Email"`
    Age       int    `quire:"Age"`
    Country   string `quire:"Country"`
    IsActive  bool   `quire:"IsActive"`
}

func getActiveUsersByCountry(ctx context.Context, db *quire.DB, country string) ([]User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Country", "=", country).
        Where("IsActive", "=", true).
        Where("Age", ">=", 18).
        Get(ctx, &users)
    return users, err
}

func getAdultUsers(ctx context.Context, db *quire.DB, limit int) ([]User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Age", ">=", 18).
        Limit(limit).
        Get(ctx, &users)
    return users, err
}

func findUserByEmail(ctx context.Context, db *quire.DB, email string) (*User, error) {
    var users []User
    err := db.Table("Users").Query().
        Where("Email", "=", email).
        Limit(1).
        Get(ctx, &users)
    if err != nil {
        return nil, err
    }
    if len(users) == 0 {
        return nil, fmt.Errorf("usuario no encontrado")
    }
    return &users[0], nil
}
```

### Uso con Context y Timeouts

```go
func getUsersWithTimeout(db *quire.DB) ([]User, error) {
    // Crear context con timeout de 5 segundos
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    var users []User
    err := db.Table("Users").Query().
        Limit(100).
        Get(ctx, &users)
    
    if err != nil {
        if errors.Is(err, context.DeadlineExceeded) {
            return nil, fmt.Errorf("timeout al consultar usuarios")
        }
        return nil, err
    }
    
    return users, nil
}
```

## Mejores Prácticas

### 1. Manejo de Errores

```go
if err != nil {
    // No usar log.Fatal en código de producción
    // Mejor retornar el error para manejarlo arriba en la pila
    return fmt.Errorf("failed to query users: %w", err)
}
```

### 2. Cierre de Conexiones

```go
db, err := quire.New(config)
if err != nil {
    return err
}
defer db.Close()  // Siempre usar defer
```

### 3. Validación de Datos

```go
// Validar antes de insertar
if user.Age < 0 {
    return fmt.Errorf("age cannot be negative")
}

if user.Email == "" {
    return fmt.Errorf("email is required")
}

err := db.Table("Users").Insert(ctx, []User{user})
```

### 4. Paginación Manual

```go
func getUsersPaginated(ctx context.Context, db *quire.DB, page, pageSize int) ([]User, error) {
    // Nota: Esto carga todos los datos y luego filtra
    // Para grandes volúmenes, considera otras estrategias
    var allUsers []User
    err := db.Table("Users").Query().Get(ctx, &allUsers)
    if err != nil {
        return nil, err
    }
    
    start := (page - 1) * pageSize
    end := start + pageSize
    
    if start >= len(allUsers) {
        return []User{}, nil
    }
    
    if end > len(allUsers) {
        end = len(allUsers)
    }
    
    return allUsers[start:end], nil
}
```

### 5. Campos Sensibles

```go
type User struct {
    ID       int    `quire:"ID"`
    Username string `quire:"Username"`
    Email    string `quire:"Email"`
    Password string `quire:"-"`  // Nunca almacenar en Sheets
}
```

### 6. Caché de Conexión

```go
type UserRepository struct {
    db *quire.DB
}

func NewUserRepository(credentials []byte, spreadsheetID string) (*UserRepository, error) {
    db, err := quire.New(quire.Config{
        SpreadsheetID: spreadsheetID,
        Credentials:   credentials,
    })
    if err != nil {
        return nil, err
    }
    
    return &UserRepository{db: db}, nil
}

func (r *UserRepository) Close() error {
    return r.db.Close()
}

func (r *UserRepository) GetUsers(ctx context.Context) ([]User, error) {
    var users []User
    err := r.db.Table("Users").Query().Get(ctx, &users)
    return users, err
}
```

## Limitaciones

1. **No hay UPDATE ni DELETE**: Quire solo soporta CREATE (Insert) y READ (Query). Para modificar o eliminar datos, debes hacerlo manualmente en Google Sheets o implementar tu propia lógica.

2. **No hay transacciones**: Las operaciones no son atómicas.

3. **Rate Limits de Google**: Google Sheets API tiene cuotas:
   - 500 requests por 100 segundos por proyecto
   - 100 requests por 100 segundos por usuario

4. **Límite de filas**: Google Sheets soporta hasta 10 millones de celdas por spreadsheet.

5. **Ordenamiento**: La función `OrderBy` actualmente es un placeholder y no ordena realmente los resultados.

6. **Concurrencia**: Aunque Quire soporta `context.Context`, no hay control de concurrencia a nivel de filas.

## Solución de Problemas

### Error: "spreadsheet ID is required"

Verifica que estás pasando el SpreadsheetID correctamente:

```go
db, err := quire.New(quire.Config{
    SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", // Sin espacios
    Credentials:   credentials,
})
```

### Error: "credentials are required"

Asegúrate de que el archivo de credenciales existe y no está vacío:

```go
credentials, err := os.ReadFile("service-account.json")
if err != nil {
    log.Fatal("No se pudo leer el archivo de credenciales:", err)
}

if len(credentials) == 0 {
    log.Fatal("El archivo de credenciales está vacío")
}
```

### Error: "failed to read data"

Posibles causas:
1. La Service Account no tiene acceso al spreadsheet
2. El spreadsheet no existe
3. La hoja (sheet) no existe

Solución:
- Verifica que compartiste el spreadsheet con el email de la Service Account
- Verifica que el nombre de la tabla coincide con el nombre de la pestaña

### Error: "googleapi: Error 403: Forbidden"

La Service Account no tiene permisos suficientes:
- Ve a Google Sheets > Share
- Asegúrate de dar permisos de "Editor"
- Verifica que no haya restricciones de dominio

### Los datos no se mapean correctamente

Verifica que:
1. Los headers en la primera fila coincidan exactamente con los tags `quire` o nombres de campo
2. Los tipos de datos sean compatibles
3. No haya espacios extra en los headers

```go
// Sheet: | ID | Name  | Email |
// Struct:  ID   Name    Email   <- Coinciden exactamente

type User struct {
    ID    int    `quire:"ID"`
    Name  string `quire:"Name"`
    Email string `quire:"Email"`
}
```

### Los filtros no funcionan

- Verifica el nombre de la columna (case-sensitive)
- Asegúrate de que los operadores sean válidos: `=`, `!=`, `>`, `>=`, `<`, `<=`, `contains`
- Para strings, usa `contains` para búsquedas parciales

### Performance lenta

Para grandes volúmenes de datos:
1. Usa `Limit()` para paginar resultados
2. Considera dividir datos en múltiples sheets
3. Implementa caché local si los datos no cambian frecuentemente
4. Ten en cuenta los rate limits de Google API

## Licencia

MIT License - ver [LICENSE](LICENSE) para más detalles.

## Contribuir

Las contribuciones son bienvenidas. Por favor:

1. Fork el repositorio
2. Crea una rama para tu feature (`git checkout -b feature/amazing-feature`)
3. Commitea tus cambios (`git commit -m 'Add amazing feature'`)
4. Push a la rama (`git push origin feature/amazing-feature`)
5. Abre un Pull Request

## Agradecimientos

- [Google Sheets API](https://developers.google.com/sheets/api)
- [Google API Go Client](https://github.com/googleapis/google-api-go-client)
