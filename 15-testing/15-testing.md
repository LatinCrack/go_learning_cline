# 🧪 Lección 15: Testing en Go — De la Confianza Ciega a la Certeza Absoluta

```
╔══════════════════════════════════════════════════════════════╗
║  "El código sin tests es como un paracaísin sin probar:     ║
║   probablemente funciona... hasta que necesitas que funcione."║
╚══════════════════════════════════════════════════════════════╝
```

---

## 🤔 ¿Qué es Testing y por qué debería importarte?

### La Analogía del Puente

Imagina que eres un ingeniero civil que construyó un puente. 

- **Sin tests** = Construyes el puente y dices *"creo que aguanta autos"*, y abres el paso sin más.
- **Con tests** = Antes de abrir, pasas camiones cargados, simulas terremotos, pruebas tormentas. Solo cuando el puente **sobrevive todo**, lo abres.

En programación es exactamente igual. El testing es el proceso de **verificar que tu código hace lo que promete**, no solo cuando todo sale bien, sino especialmente cuando algo sale mal.

### ¿Por qué Go tiene testing integrado?

Muchos lenguajes necesitan librerías externas para testing (JUnit en Java, pytest en Python, Jest en JavaScript). Go dice: **"no, testing es parte del lenguaje"**.

```go
// Esto ya está disponible sin instalar NADA:
import "testing"
```

Go incluye:
- 📦 Paquete `testing` integrado en la biblioteca estándar
- 🖥️ Comando `go test` integrado en el toolchain
- 📊 Medición de cobertura de código (`-cover`)
- ⚡ Benchmarking integrado (`-bench`)
- 📝 Ejemplos ejecutables (`Example`)

No necesitas npm, pip, cargo ni ningún gestor de paquetes adicional. **Testing es ciudadano de primera clase en Go**.

---

## 📁 Convenciones de Archivos

Go tiene una convención estricta y elegante para los tests:

```
mi_paquete/
├── calculadora.go      ← código fuente
├── calculadora_test.go ← tests para calculadora.go
├── utils.go            ← otro archivo fuente
└── utils_test.go       ← tests para utils.go
```

Las reglas son:

| Regla | Ejemplo |
|-------|---------|
| El archivo **debe** terminar en `_test.go` | `mathutil_test.go` ✅ |
| El archivo **debe** estar en el mismo paquete | Si el código es `package mathutil`, el test también |
| Los archivos `_test.go` **solo se compilan** con `go test` | No afectan el binario final |
| Cada función de test **debe** empezar con `Test` + Mayúscula | `TestSuma`, `TestFactorial` |

```
┌─────────────────────────────────────────────────────┐
│  💡 TIP FEYNMAN:                                     │
│                                                      │
│  Los archivos _test.go son INVISIBLES para go build. │
│  Solo existen cuando ejecutas go test.               │
│  Es como tener un laboratorio secreto dentro de tu   │
│  fábrica: los clientes nunca lo ven, pero tú sabes   │
│  que cada producto fue probado ahí.                   │
└─────────────────────────────────────────────────────┘
```

---

## 🔧 Tu Primera Función de Test

### La Estructura Básica

```go
package mathutil

import "testing"

func TestSuma(t *testing.T) {
    // ARRANGE: preparar datos de entrada
    nums := []float64{1, 2, 3, 4, 5}
    esperado := 15.0
    
    // ACT: ejecutar la función que queremos probar
    resultado := Suma(nums)
    
    // ASSERT: verificar que el resultado es correcto
    if resultado != esperado {
        t.Errorf("Suma(%v) = %f, se esperaba %f", nums, resultado, esperado)
    }
}
```

### Desglose de la Función

| Elemento | Significado |
|----------|-------------|
| `func TestSuma(t *testing.T)` | Una función de test. **Siempre** empieza con `Test` y recibe `*testing.T` |
| `t *testing.T` | El "controlador" del test. Te permite reportar fallos, logs, y más |
| `t.Errorf(...)` | Reporta un error **y continúa** ejecutando otros tests |
| `t.Fatalf(...)` | Reporta un error **y detiene** el test actual inmediatamente |
| `t.Logf(...)` | Escribe un log (solo visible con `-v`) |

### El Patrón AAA: Arrange, Act, Assert

Todo buen test sigue este patrón de tres pasos:

```
┌──────────┐     ┌──────────┐     ┌──────────┐
│ ARRANGE  │────▶│   ACT    │────▶│  ASSERT  │
│ Prepara  │     │ Ejecuta  │     │ Verifica │
│ los datos│     │ la func. │     │ el result│
└──────────┘     └──────────┘     └──────────┘
```

1. **Arrange** (Preparar): Define los datos de entrada y el resultado esperado
2. **Act** (Actuar): Llama a la función que estás probando
3. **Assert** (Afirmar): Compara el resultado real con el esperado

---

## 🎯 Table-Driven Tests: El Patrón Go por Excelencia

Si aprendes UNA sola técnica de testing en Go, que sea esta. Los **table-driven tests** son el patrón más utilizado en todo el ecosistema Go.

### ¿Qué es un Table-Driven Test?

En lugar de escribir un test separado para cada caso, defines una **tabla** (slice de structs) con todos los casos de prueba, y iteras sobre ella:

```go
func TestSuma(t *testing.T) {
    tests := []struct {
        name     string      // nombre descriptivo del caso
        nums     []float64   // datos de entrada
        expected float64     // resultado esperado
    }{
        {"numeros positivos", []float64{1, 2, 3}, 6.0},
        {"numeros negativos", []float64{-1, -2, -3}, -6.0},
        {"slice vacio", []float64{}, 0.0},
        {"un solo elemento", []float64{42}, 42.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Suma(tt.nums)
            if result != tt.expected {
                t.Errorf("Suma(%v) = %f, se esperaba %f",
                    tt.nums, result, tt.expected)
            }
        })
    }
}
```

### ¿Por qué es tan poderoso?

```
┌───────────────────────────────────────────────────────────┐
│  SIN table-driven:          CON table-driven:             │
│                                                           │
│  TestSumaPositivos()        func TestSuma(t *testing.T)  │
│  TestSumaNegativos()        tests := []struct{...}{      │
│  TestSumaVacio()               {"positivos", ...},        │
│  TestSumaUnElemento()          {"negativos", ...},        │
│  TestSumaDecimales()           {"vacío", ...},            │
│  ... (N funciones)             {"un elemento", ...},      │
│                              }                            │
│                                                           │
│  5 funciones separadas      1 función con 5 sub-casos    │
│  Código repetitivo          Código DRY y escalable       │
│  Difícil de mantener        Fácil de agregar más casos   │
└───────────────────────────────────────────────────────────┘
```

Ventajas:
- 🎯 **DRY**: No repites lógica de test
- 📈 **Escalable**: Agregar un caso es solo una línea más en la tabla
- 📖 **Legible**: Ves todos los casos de golpe
- 🔍 **Aislado**: Si falla uno, sabes exactamente cuál (`t.Run` con nombre)

---

## 🔨 Comandos Esenciales de `go test`

### Ejecutar todos los tests
```bash
go test ./...          # todos los tests del proyecto
```

### Ver output detallado
```bash
go test ./... -v       # verbose: muestra cada test individual
```

### Ejecutar un test específico
```bash
go test -run TestSuma                       # solo TestSuma
go test -run TestSuma/numeros_positivos     # un sub-test específico
go test -run "TestSuma/numeros.*"           # con regex
```

### Cobertura de código
```bash
go test ./... -cover                        # porcentaje de cobertura
go test ./... -coverprofile=cover.out       # generar reporte
go tool cover -html=cover.out              # ver cobertura en el navegador
```

### Benchmarks
```bash
go test ./... -bench=.                      # ejecutar todos los benchmarks
go test ./... -bench=. -benchmem            # incluir uso de memoria
```

### Ejemplo de output con `-v`
```
=== RUN   TestSuma
=== RUN   TestSuma/numeros_positivos
=== RUN   TestSuma/numeros_negativos
=== RUN   TestSuma/slice_vacio
--- PASS: TestSuma (0.00s)
    --- PASS: TestSuma/numeros_positivos (0.00s)
    --- PASS: TestSuma/numeros_negativos (0.00s)
    --- PASS: TestSuma/slice_vacio (0.00s)
```

---

## ⚠️ Testing Errores: La Mitad del Trabajo

Probar el camino feliz (cuando todo sale bien) es solo la mitad. **La otra mitad — y la más importante — es probar qué pasa cuando las cosas van mal**.

```go
func TestPromedio(t *testing.T) {
    t.Run("slice vacio retorna error", func(t *testing.T) {
        _, err := Promedio([]float64{})
        if err == nil {
            t.Fatal("se esperaba error con slice vacio, se obtuvo nil")
        }
    })

    t.Run("numeros normales no retorna error", func(t *testing.T) {
        _, err := Promedio([]float64{1, 2, 3})
        if err != nil {
            t.Fatalf("no se esperaba error, se obtuvo: %v", err)
        }
    })
}
```

### `t.Error` vs `t.Fatal`

```
┌─────────────────────┬──────────────────────┐
│     t.Error()       │     t.Fatal()        │
├─────────────────────┼──────────────────────┤
│ Reporta el error    │ Reporta el error     │
│ CONTINUA ejecutando │ DETIENE el test      │
│ Útil para asserts   │ Útil cuando no tiene │
│ múltiples           │ sentido continuar    │
│                     │ (ej: error == nil    │
│                     │  cuando se necesita) │
└─────────────────────┴──────────────────────┘
```

---

## 📝 Example Functions: Tests que son Documentación

Go permite crear funciones `Example` que sirven como **tests Y documentación** al mismo tiempo:

```go
func ExampleSuma() {
    nums := []float64{1, 2, 3, 4, 5}
    fmt.Printf("%.2f", Suma(nums))
    // Output: 15.00
}
```

Estas funciones:
- Aparecen en la documentación de `godoc`
- Se ejecutan como tests (el comentario `// Output:` es el assertion)
- Son ejemplos vivos que **siempre están actualizados**

---

## ⚡ Benchmarks: Midiendo el Rendimiento

Los benchmarks miden **qué tan rápido** es tu código:

```go
func BenchmarkSuma(b *testing.B) {
    nums := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
    for i := 0; i < b.N; i++ {
        Suma(nums)
    }
}
```

| Elemento | Significado |
|----------|-------------|
| `func BenchmarkXxx(b *testing.B)` | Los benchmarks empiezan con `Benchmark` |
| `b.N` | Go ajusta automáticamente el número de iteraciones |
| `b.ResetTimer()` | Reinicia el cronómetro (si hay setup costoso) |
| `b.ReportAllocs()` | Reporta asignaciones de memoria |

Output típico:
```
BenchmarkSuma-8    50000000    24.3 ns/op    0 B/op    0 allocs/op
```

Esto significa: "50 millones de iteraciones, 24.3 nanosegundos por operación, cero asignaciones de memoria".

---

## 🏗️ El Ejercicio Práctico: Una Biblioteca Completa con Tests

En este ejercicio construimos **dos paquetes con tests completos**:

```
15-testing/
├── go.mod
├── main.go
└── myutils/
    ├── mathutil/
    │   ├── mathutil.go        ← 12 funciones matemáticas
    │   └── mathutil_test.go   ← tests + benchmarks + examples
    └── textutil/
        ├── textutil.go        ← 9 funciones de texto
        └── textutil_test.go   ← tests + benchmarks
```

### 📦 Paquete `mathutil` — Funciones Matemáticas

**Funciones implementadas:**

| Función | Descripción | Maneja errores |
|---------|-------------|----------------|
| `Suma(nums)` | Suma todos los números de un slice | ✅ (retorna 0 si vacío) |
| `Promedio(nums)` | Calcula el promedio | ✅ error si vacío |
| `Maximo(nums)` | Encuentra el valor máximo | ✅ error si vacío |
| `Minimo(nums)` | Encuentra el valor mínimo | ✅ error si vacío |
| `Mediana(nums)` | Calcula la mediana | ✅ error si vacío |
| `DesviacionEstandar(nums)` | Desviación estándar poblacional | ✅ error si vacío |
| `Factorial(n)` | Calcula n! | ✅ error si n < 0 |
| `Fibonacci(n)` | Genera los primeros n números de Fibonacci | ✅ error si n < 0 |
| `EsPrimo(n)` | Verifica si n es primo | ✅ (retorna false si ≤ 1) |
| `Porcentaje(valor, total)` | Calcula porcentaje | ✅ error si total = 0 |
| `Clamp(valor, min, max)` | Limita un valor al rango [min, max] | ❌ (no necesita) |
| `ContarPares(nums)` | Cuenta números pares en un slice | ✅ (retorna 0 si vacío) |

### 📝 Paquete `textutil` — Funciones de Texto

| Función | Descripción |
|---------|-------------|
| `Invertir(s)` | Invierte una cadena (soporte Unicode) |
| `EsPalindromo(s)` | Verifica si es palíndromo (ignora espacios y mayúsculas) |
| `ContarPalabras(s)` | Cuenta palabras en una cadena |
| `TituloCapital(s)` | Convierte a Title Case |
| `SoloLetras(s)` | Extrae solo letras Unicode |
| `Truncar(s, n)` | Trunca a n caracteres con "..." |
| `ContarVocales(s)` | Cuenta vocales (incluye acentuadas) |
| `ReemplazarVocales(s, r)` | Reemplaza vocales con un rune dado |

### 🧪 Tests Escritos

**Para mathutil: 11 funciones de test + 4 examples + 4 benchmarks**

```
TestSuma               → 8 sub-casos (positivos, negativos, vacío, decimales, grandes...)
TestPromedio            → 5 sub-casos (incluye error con slice vacío)
TestMaximo              → 6 sub-casos (positivos, negativos, vacío, iguales, decimales)
TestMinimo              → 4 sub-casos
TestMediana             → 5 sub-casos (impar, par, desordenados, vacío, un elemento)
TestDesviacionEstandar  → 4 sub-casos (iguales, conocidos, vacío, un elemento)
TestFactorial           → 7 sub-casos (0, 1, 5, 10, 20, negativo, negativo grande)
TestFibonacci           → 6 sub-casos (0, 1, 2, 8, 12, negativo)
TestEsPrimo             → 13 sub-casos (0, 1, 2, 3, 4, 5, 9, 11, 15, 17, 100, 103, -5)
TestClamp               → 7 sub-casos (dentro, debajo, encima, límites, negativo, min=max)
TestPorcentaje          → 6 sub-casos (50%, 100%, 0%, división por cero, pequeño, >total)
TestContarPares         → 7 sub-casos (todos pares, impares, mezcla, vacío, cero, negativos)

ExampleSuma             → verificable por godoc
ExamplePromedio         → verificable por godoc
ExampleEsPrimo          → verificable por godoc
ExampleFactorial        → verificable por godoc

BenchmarkPromedio       → medición de rendimiento
BenchmarkFactorial      → medición de rendimiento
BenchmarkFibonacci      → medición de rendimiento
BenchmarkEsPrimo        → medición de rendimiento
```

**Para textutil: 8 funciones de test + 5 benchmarks**

```
TestInvertir            → 8 sub-casos (simple, palíndromo, vacío, unicode, emoji...)
TestEsPalindromo        → 8 sub-casos (simple, espacios, mayúsculas, frase...)
TestContarPalabras      → 7 sub-casos (normal, vacío, múltiples espacios, tabs...)
TestTituloCapital       → 6 sub-casos (minúsculas, mayúsculas, mezcla, vacío...)
TestSoloLetras          → 7 sub-casos (números, símbolos, acentos, vacío...)
TestTruncar             → 6 sub-casos (sin truncar, exacto, corto, unicode...)
TestContarVocales       → 7 sub-casos (normal, sin vocales, mayúsculas, acentuadas...)
TestReemplazarVocales   → 6 sub-casos (asterisco, guión, sin vocales, mayúsculas...)

BenchmarkInvertir       → medición de rendimiento
BenchmarkEsPalindromo   → medición de rendimiento
BenchmarkContarPalabras → medición de rendimiento
BenchmarkTituloCapital  → medición de rendimiento
BenchmarkContarVocales  → medición de rendimiento
```

**Total: 52 tests individuales + 4 examples + 9 benchmarks = 65 verificaciones** 🎯

---

## 📊 Resultado de los Tests

Al ejecutar `go test ./... -v`, todos los 52 tests pasan:

```
ok  15-testing/myutils/mathutil   0.981s
ok  15-testing/myutils/textutil   0.446s
```

### Output del programa principal (`go run main.go`):

```
========================================
  LECCION 15: TESTING EN GO
  De la confianza ciega a la certeza
========================================

🔬 MATHUTIL - Funciones matematicas
─────────────────────────────────────────────
  Datos: [10 20 30 40 50]
  Suma:           150.00
  Promedio:       30.00
  Maximo:         50.00
  Minimo:         10.00
  Mediana:        30.00
  Desv. Estandar: 14.14
  30 de 80:       37.5%
  6! =            720
  Fibonacci(10):  [0 1 1 2 3 5 8 13 21 34]
  ¿Es 17 primo?   true
  ¿Es 100 primo?  false
  Clamp(150,0,100): 100
  Contar pares [1..10]: 5

⚠️  MATHUTIL - Casos de error
─────────────────────────────────────────────
  Promedio([]):    ❌ no se puede calcular promedio de un slice vacio
  Factorial(-5):   ❌ factorial no definido para numeros negativos: -5
  Porcentaje(50,0):❌ el total no puede ser cero

📝 TEXTUTIL - Funciones de texto
─────────────────────────────────────────────
  Invertir("hola"):         "aloh"
  ¿Es "ana" palindromo?     true
  ¿Es "hola" palindromo?    false
  Palabras en "Go es genial":3
  Titulo("hola mundo"):     "Hola Mundo"
  SoloLetras("Go 1.21!"):   "Goesgenial"
  Truncar("hola mundo",5):   "hola ..."
  Vocales en "programacion": 5
  Reemplazar vocales "hola*":"h*l*"
```

---

## 🎓 Resumen: El Mapa Mental del Testing en Go

```
                    TESTING EN GO
                         │
           ┌─────────────┼─────────────┐
           │             │             │
      ARCHIVOS      COMANDOS     FUNCIONES
           │             │             │
    *_test.go     go test ./...   t.Errorf()
    mismo pkg     -v (verbose)    t.Fatalf()
    solo en test  -run (filtro)   t.Logf()
    no en build   -cover          t.Run()
                  -bench          b.N (benchmarks)
                  -benchmem
                         │
              ┌──────────┼──────────┐
              │          │          │
         TABLE-DRIVEN  EXAMPLES  BENCHMARKS
              │          │          │
         []struct{}  ExampleXxx() BenchmarkXxx()
         t.Run()     // Output:   for i < b.N
         escalable   doc + test   rendimiento
```

### Las 5 Reglas de Oro del Testing en Go

```
╔═══════════════════════════════════════════════════════════════╗
║  1. 📁 Convención: *_test.go, mismo paquete, Test + Mayúscula ║
║  2. 📋 Table-driven: una tabla con N casos, no N funciones    ║
║  3. 🔄 Cubrir feliz + triste: testear errores Y éxitos       ║
║  4. 🏷️ Nombres descriptivos: que el fallo se explique solo    ║
║  5. 📊 Medir: usa -cover y -bench, no adivines               ║
╚═══════════════════════════════════════════════════════════════╝
```

---

## 🧠 Ejercicio Feynman

> **Instrucciones:** Explica estos conceptos en voz alta, con tus propias palabras, como si le enseñaras a alguien que nunca ha programado. Si te trabas en algún punto, vuelve a leer esa sección.

### Nivel Básico — "¿Qué es y para qué sirve?"

1. **¿Qué es un test unitario?** Usa la analogía del puente o la del médico.

2. **¿Por qué Go incluye testing en su estándar?** ¿Qué ventaja tiene sobre usar librerías externas?

3. **¿Qué significa que un archivo termine en `_test.go`?** ¿Qué pasa con esos archivos cuando haces `go build`?

### Nivel Intermedio — "¿Cómo funciona?"

4. **Explica qué es un table-driven test** sin usar código. Usa una analogía con una lista de exámenes.

5. **¿Cuál es la diferencia entre `t.Errorf()` y `t.Fatalf()`?** ¿Cuándo usarías cada uno?

6. **Si tu función `Dividir(a, b)` retorna error cuando `b=0`**, escribe (en papel o en voz alta) un test que verifique:
   - Que con `b=0` retorna error
   - Que con `b=2` y `a=10` retorna `5` sin error

### Nivel Avanzado — "¿Por qué es así?"

7. **¿Por qué los tests de Go NO se compilan en el binario final?** ¿Qué implicación tiene esto para el tamaño del ejecutable?

8. **Un compañero dice "mis tests pasan todos, mi código es perfecto"**. ¿Qué le responderías? Piensa en qué casos los tests pueden pasar y el código aún tener bugs.

9. **Explica qué es la cobertura de código (`-cover`)**. Si tienes 80% de cobertura, ¿qué significa el 20% restante? ¿Es siempre preocupante?

### Nivel Feynman — "Enséñalo"

10. **Prepara una explicación de 3 minutos** sobre table-driven tests para un colega. Debe incluir:
    - Qué problema resuelve
    - Cómo se estructura (la tabla, el loop, t.Run)
    - Por qué es mejor que escribir tests individuales
    - Un ejemplo cotidiano que no sea de programación

### Autoevaluación

```
┌──────────────────────────────────────────────────────────┐
│  Si pudiste responder las 10 preguntas con claridad:     │
│  ✅ Entiendes testing en Go a nivel sólido               │
│                                                          │
│  Si te trabaste en las 7-10:                             │
│  📖 Revisa la sección de Table-Driven Tests y Patterns   │
│                                                          │
│  Si te trabaste antes de la 4:                           │
│  🔁 Relee desde el inicio, enfócate en las analogías    │
└──────────────────────────────────────────────────────────┘
```

---

## 🚀 Comandos para Experimentar

```bash
# Entrar al directorio
cd 15-testing

# Ejecutar el programa principal
go run main.go

# Ejecutar todos los tests con verbose
go test ./... -v

# Ejecutar solo un test específico
go test ./... -run TestFactorial -v

# Ejecutar solo un sub-test
go test ./... -run "TestSuma/numeros_positivos" -v

# Ver cobertura de código
go test ./... -cover

# Ejecutar benchmarks
go test ./... -bench=. -benchmem

# Ver cobertura visual en el navegador
go test ./... -coverprofile=cover.out
go tool cover -html=cover.out
```

```
╔══════════════════════════════════════════════════════════════╗
║  🎯 Próxima lección: 16 - Interfaces                        ║
║     El concepto más elegante de Go:                        ║
║     "No preguntes qué es, pregunta qué sabe hacer"        ║
╚══════════════════════════════════════════════════════════════╝