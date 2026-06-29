<div align="center">

# 🚀 Laboratorio de Go: De Cero a Experto

### *Método Feynman · 20 Lecciones · Pensamiento Lateral*

<br>

![Go](https://img.shields.io/badge/Language-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white) ![Método Feynman](https://img.shields.io/badge/M%C3%A9todo-Feynman-FF6B6B?style=for-the-badge) ![20 Lecciones](https://img.shields.io/badge/Lecciones-20-4ECDC4?style=for-the-badge) ![De Cero a Experto](https://img.shields.io/badge/Nivel-Cero_a_Experto-FFE66D?style=for-the-badge)

<br>

> *"Si no puedes explicarlo de forma simple, no lo entiendes bien."*
> — **Richard Feynman**

</div>

---

## 📖 Acerca de Esta Ruta de Aprendizaje

Este laboratorio está diseñado bajo el **Método Feynman**: aprender un concepto, luego intentar enseñarlo con palabras propias como si se lo explicaras a alguien sin contexto técnico. Cada lección combina **teoría profunda**, **ejercicios prácticos del mundo real** y un **Feynman Challenge** que pondrá a prueba tu comprensión real.

### 🎯 Filosofía del Curso

| Principio | Descripción |
|:----------|:------------|
| 🧠 **Aprender haciendo** | Cada concepto se ancla con un ejercicio útil, no ejercicios académicos inventados |
| 🔄 **Spiral Learning** | Los temas se retoman y profundizan en lecciones posteriores |
| 🌍 **Contexto real** | Todos los ejercicios resuelven problemas reales de ingeniería |
| 🪶 **Simplificación forzada** | El Feynman Challenge expone tus puntos ciegos |

---

## 🗺️ Mapa de Ruta Completo

<br>

### 📗 FASE I — Fundamentos (Lecciones 01–05)
> *Construye los cimientos. Aquí nace tu intuición en Go.*

---

#### 📘 [Lección 01 — Hello, Go! Entorno, Sintaxis y el Pulso del Lenguaje](./01-hello-go/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Comprender la filosofía de Go como lenguaje: simplicidad como superpoder, compilación rápida, y por qué Go fue diseñado para resolver los dolores de cabeza de la ingeniería en Google (compilaciones lentas, dependencias complejas, concurrencia frágil).
- **Analogía mental:** Go es como una navaja suiza minimalista. No tiene 47 herramientas como un lenguaje multi-paradigma, pero las que tiene están afiladas quirúrgicamente. Aprender Go es como aprender a usar un set de herramientas de cirujano: pocas, precisas, devastadoramente eficientes.
- **Caso de uso:** Entender cómo Go compila a un solo binario estático sin dependencias — esto es lo que hace que herramientas como `Docker`, `Kubernetes` y `Terraform` sean tan fáciles de distribuir.

**🏋️ Ejercicio Práctico: Configuración del Laboratorio y Tu Primer CLI**

Construirás un script CLI mínimo en Go que funcione como un **detector de entorno de desarrollo**: al ejecutarlo, mostrará tu sistema operativo, arquitectura, versión de Go, y el `GOPATH` configurado. Este será tu "panel de instrumentos" del laboratorio.

> **¿Por qué este enfoque?** Porque antes de construir una casa necesitas verificar que tus cimientos están bien. Este CLI será la base que usarás para diagnosticar problemas a lo largo de todo el curso.

```bash
# Estructura del proyecto
01-hello-go/
├── main.go
└── go.mod
```

**🧠 Feynman Challenge**

> Imagina que un amigo que trabaja en marketing te pregunta: *"¿Para qué necesito otro lenguaje si ya existen Python y JavaScript?"*. Explícale en 5 oraciones por qué Go existe, qué problema resuelve, y cuándo NO debería usarlo. Si te trabas en el "cuándo no usarlo", es porque todavía no entiendes bien el problema que Go resuelve.

</details>

---

#### 📘 [Lección 02 — Variables, Tipos y el Sistema de Tipos de Go](./02-variables/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Dominar el sistema de tipos estáticos de Go, entender la diferencia entre tipos primitivos (`int`, `float64`, `string`, `bool`), tipos compuestos (`array`, `slice`, `map`), y por qué Go no tiene herencia ni clases — tiene algo mejor.
- **Analogía mental:** Los tipos en Go son como las etiquetas de los ingredientes en un restaurante profesional. Si el chef pide "harina", no quiere que le des "azúcar" aunque ambos sean polvos blancos. El compilador de Go es ese chef exigente que no te deja mezclar ingredientes incorrectos.
- **Caso de uso:** El sistema de tipos de Go previene bugs silenciosos que en lenguajes dinámicos solo aparecen en producción a las 3 AM.

**🏋️ Ejercicio Práctico: Conversor de Unidades Universal**

Construirás una herramienta CLI que convierte unidades del mundo real: temperaturas (°C ↔ °F ↔ K), distancias (km ↔ millas ↔ metros), y almacenamiento (bytes ↔ KB ↔ MB ↔ GB). La herramienta leerá argumentos desde la línea de comandos y mostrará todas las conversiones posibles.

> **¿Por qué este enfoque?** Porque un conversor de unidades es el ejemplo perfecto para dominar tipos numéricos, conversiones explícitas (Go no hace casting automático), y el formateo de strings con `fmt.Sprintf`. Además, es una herramienta que usarás en la vida real.

```bash
go run main.go temp 100
# 100°C = 212°F = 373.15K
```

**🧠 Feynman Challenge**

> Go tiene tipos estáticos como Java, pero la gente dice que es más fácil. ¿Por qué? Explica con tus propias palabras qué significa que Go tenga inferencia de tipos (`:=`), por qué Go no permite conversiones implícitas (`int` a `float64`), y qué pasaría si el compilador te dejara sumar un `string` con un `int`. Dibuja mentalmente el escenario donde eso causa un desastre en producción.

</details>

---

#### 📘 [Lección 03 — Tipos de Datos en Go](./03-tipos-de-datos/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go solo tiene `for` como bucle (no hay `while`, no hay `do-while`). Entender por qué esta decisión radical de diseño hace el código más legible, y cómo el `switch` de Go es más poderoso que en C/Java porque no necesita `break` y puede evaluar cualquier tipo.
- **Analogía mental:** Si los bucles fueran utensilios de cocina, otros lenguajes te dan 5 tipos de cuchillos diferentes. Go te da uno solo (el `for`), pero ese cuchillo corta TODO — desde pan hasta diamantes. La simplicidad no es debilidad, es maestría.
- **Caso de uso:** Los servidores HTTP y pipelines de procesamiento usan `for` con `select` para manejar miles de conexiones simultáneas — el mismo `for` que aprenderás aquí.

**🏋️ Ejercicio Práctico: Analizador de Logs del Sistema**

Construirás un parser de archivos de log que lea líneas de texto, identifique patrones (errores, warnings, info), cuente frecuencias, y genere un reporte de resumen. Usará `for` para iterar, `switch` para clasificar niveles de severidad, e `if` para filtrar por fechas.

> **¿Por qué este enfoque?** Porque el parsing de logs es una de las tareas más comunes en operaciones de software. Este ejercicio te obliga a usar TODAS las estructuras de control de Go en un contexto real y medible.

```bash
go run main.go /var/log/app.log
# 📊 Reporte: 1,247 errores | 3,891 warnings | 89,234 info
# 🔴 Error más frecuente: "connection timeout" (342 veces)
```

**🧠 Feynman Challenge**

> Explícale a un niño de 12 años por qué Go solo tiene un tipo de bucle (`for`) mientras que otros lenguajes tienen `for`, `while`, `do-while`, `foreach`, etc. Usa la analogía de una caja de herramientas. Luego intenta explicar el `switch` de Go sin usar la palabra "case" — si no puedes, es porque estás memorizando sintaxis en vez de entender el concepto.

</details>

---

#### 📘 [Lección 04 — Funciones: El ADN de Go](./04-funciones/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Las funciones en Go son ciudadanos de primera clase: pueden ser asignadas a variables, pasadas como argumentos, y devueltas desde otras funciones. Go soporta múltiples valores de retorno, funciones variádicas, y closures — todo sin la complejidad de los genéricos de Java o la ambigüedad de JavaScript.
- **Analogía mental:** Una función en Go es como una receta de cocina profesional. Tiene ingredientes claros (parámetros tipados), produce platos definidos (valores de retorno nombrados), y puede incluso devolver un "error" como segundo plato si algo sale mal. Esto es único: en Go, los errores son valores, no excepciones.
- **Caso de uso:** El patrón `(resultado, error)` de Go es el estándar de la industria. Desde la librería estándar hasta proyectos de Kubernetes, TODO en Go devuelve errores como valores — no como excepciones que explotan en tu cara.

**🏋️ Ejercicio Práctico: Calculadora de Expresiones con Manejo de Errores**

Construirás una calculadora que parsea expresiones matemáticas desde texto (`"3 + 5 * (2 - 1)"`), las evalúa respetando precedencia de operadores, y maneja errores elegantemente: división por cero, paréntesis desbalanceados, caracteres inválidos. Cada función devolverá `(resultado, error)` al estilo Go puro.

> **¿Por qué este enfoque?** Porque este ejercicio te obliga a escribir funciones puras, usar closures para el parser, devolver errores como valores, y experimentar con funciones como ciudadanos de primera clase. Es el ejercicio definitivo para internalizar el manejo de errores de Go.

**🧠 Feynman Challenge**

> En Python y JavaScript, los errores se lanzan con `throw` y se capturan con `try/catch`. Go no tiene excepciones. Explica con tus propias palabras POR QUÉ Go eligió el patrón `(resultado, error)`, qué ventaja tiene sobre las excepciones, y en qué escenario las excepciones serían mejores. Si solo ves ventajas en Go, no has pensado lo suficiente.

</details>

---

#### 📘 [Lección 05 — Estructuras, Métodos y la Filosofía "Composición sobre Herencia"](./05-structs/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go no tiene clases ni herencia. En su lugar, usa **structs** para agrupar datos y **métodos** para comportamiento. La composición mediante **embedding** (structs dentro de structs) reemplaza la herencia, resultando en arquitecturas más flexibles y testeables.
- **Analogía mental:** Piensa en la herencia como una familia real: si el rey tiene un defecto, todos los herederos lo arrastran. La composición es como un equipo de LEGO: cada pieza es independiente, reemplazable, y puedes construir lo que quieras sin que la pieza del caballo arrastre los defectos de la pieza del barco.
- **Caso de uso:** Docker está construido con este patrón. Sus tipos de contenedores, imágenes y redes son structs que se componen entre sí, no clases que heredan.

**🏋️ Ejercicio Práctico: Mini Gestor de Contactos en Archivo**

Construirás un gestor de contactos CLI que almacena datos en un archivo JSON. Usarás structs para `Contacto`, `Agenda`, y `CriterioBusqueda`. Implementarás métodos para agregar, buscar, filtrar y exportar contactos. La agenda compondrá contactos mediante embedding.

> **¿Por qué este enfoque?** Porque un gestor de contactos requiere modelar entidades del mundo real con structs, asociarles comportamiento con métodos, y demostrar la composición. Es un dominio familiar que te permite enfocarte en la arquitectura en vez del problema de negocio.

**🧠 Feynman Challenge**

> Imagina que le explicas a un arquitecto de software de Java por qué Go no necesita `extends`, `abstract`, `virtual`, ni `interface implements`. Usa la analogía del equipo LEGO vs la familia real. Si encuentras diciendo *"pero sin herencia es más difícil..."*, detente y busca 3 razones por las que la composición es más flexible. Tu punto débil está ahí.

</details>

---

### 📘 FASE II — Tipos de Datos y Estructuras (Lecciones 06–09)
> *Domina las herramientas de datos de Go. Aquí es donde Go brilla.*

---

#### 📘 [Lección 06 — Arrays, Slices y el Secreto del Runtime de Go](./06-slices/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Los arrays en Go son de tamaño fijo y se pasan por valor (se copian). Los slices son la estrella: son ventanas dinámicas sobre un array subyacente, con un puntero, longitud y capacidad. Entender el mecanismo interno de `append`, `copy` y el crecimiento del slice es crítico para escribir Go de alto rendimiento.
- **Analogía mental:** Un array es como un estacionamiento con exactamente 10 espacios — si necesitas 11, debes construir otro estacionamiento completo. Un slice es como un auto con un remolque expandible: arranca con 3 cajones, pero si necesitas más, el runtime de Go te consigue un remolque más grande automáticamente (y mueve tus cajones).
- **Caso de uso:** El 90% del código Go usa slices, no arrays. Entender cuándo un slice comparte memoria con otro (aliasing) es el difference entre un bug de datos corruptos y código correcto.

**🏋️ Ejercicio Práctico: Procesador de Archivos CSV de Alta Velocidad**

Construirás un procesador que lee archivos CSV grandes (millones de líneas), filtra filas según criterios, agrupa datos por columnas, y calcula estadísticas (promedio, mediana, desviación estándar). Todo usando slices eficientes con `append`, slicing, y `copy`.

> **¿Por qué este enfoque?** Porque procesar CSVs es una tarea diaria en data engineering y backend. Este ejercicio te obliga a entender el crecimiento de slices, el aliasing de memoria, y la diferencia entre `append` y `copy` en un contexto donde un error de rendimiento te cuesta minutos de procesamiento.

**🧠 Feynman Challenge**

> Explica con tus propias palabras qué pasa internamente cuando haces `append(slice, elemento)` y el slice ya no tiene capacidad. Dibuja mentalmente el proceso: ¿cuándo se crea un nuevo array? ¿cuánto crece? ¿qué pasa con el slice original? Si no puedes dibujar el proceso, no entiendes slices — solo los usas.

</details>

---

#### 📘 [Lección 07 — Maps: Diccionarios, Tablas Hash y el Arte del Acceso O(1)](./07-maps/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Los maps en Go son implementaciones de tablas hash que ofrecen acceso O(1) promedio. Son referencias (no se copian), no son concurrent-safe (necesitas `sync.RWMutex` o `sync.Map`), y entender el `comma ok idiom` es esencial para distinguir entre "la clave no existe" y "el valor es el zero value".
- **Analogía mental:** Un map es como el índice de un libro gigante. Si quieres saber en qué página está "algoritmo", no lees todo el libro — vas al índice y en O(1) encuentras la página. Pero si dos personas están escribiendo en el índice al mismo tiempo (concurrencia), se arma un lío — por eso necesitas un "semaforito" (`mutex`).
- **Caso de uso:** Cachés en memoria, contadores de frecuencia, deduplicación de datos, y routing de URLs en servidores web usan maps extensivamente.

**🏋️ Ejercicio Práctico: Contador de Frecuencia de Palabras en Texto Real**

Construirás una herramienta que toma un texto (un libro, un artículo, o la salida de un comando), cuenta la frecuencia de cada palabra, ignora stopwords (artículos, preposiciones), y muestra las top-N palabras más frecuentes. El resultado se exporta a un archivo de texto formateado.

> **¿Por qué este enfoque?** Porque el análisis de frecuencia de palabras es la base del procesamiento de lenguaje natural (NLP) y los motores de búsqueda. Este ejercicio usa maps como el pilar central y te expone al `comma ok idiom`, iteración con `range`, y ordenamiento de maps (que no es trivial en Go).

**🧠 Feynman Challenge**

> Los maps de Go no son concurrent-safe. Explica por qué usando la analogía de dos personas escribiendo en el mismo pizarrón al mismo tiempo. Luego explica la diferencia entre `sync.Mutex` y `sync.RWMutex` — si solo dices "uno es más rápido", no has entendido el concepto. ¿En qué escenario usarías cada uno?

</details>

---

#### 📘 [Lección 08 — Strings, Runes y el Universo Unicode de Go](./08-strings/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Los strings en Go son **inmutables** y están codificados en **UTF-8**. Un `string` es un slice de bytes, pero un carácter Unicode puede ocupar de 1 a 4 bytes. El tipo `rune` (alias de `int32`) representa un punto de código Unicode. Esto significa que `len("ñ")` devuelve 2, no 1 — y entender esto te salvará de bugs catastróficos en aplicaciones internacionales.
- **Analogía mental:** Un `string` de Go es como un tren de vagones donde cada vagón puede ser de tamaño diferente (1-4 bytes). `len()` cuenta los vagones en metros (bytes), pero si quieres contar pasajeros (caracteres), necesitas contar personas (`runes`). Una "ñ" es un pasajero que ocupa 2 metros de vagón.
- **Caso de uso:** Cualquier aplicación que maneje texto en español, chino, árabe o emojis necesita entender runes. Un sistema de truncamiento de texto que corte a 10 "caracteres" sin entender runes cortará emojis por la mitad o separará diéresis de sus vocales.

**🏋️ Ejercicio Práctico: Truncador de Texto Unicode-Safe y Analizador de Emojis**

Construirás una herramienta que: (1) trunca textos a N caracteres visibles sin romper emojis ni acentos, (2) cuenta caracteres visibles vs bytes, (3) detecta y extrae emojis de un texto, (4) invierte strings respetando Unicode.

> **¿Por qué este enfoque?** Porque casi todos los lenguajes fallan miserablemente al manejar Unicode correctamente. Este ejercicio te convierte en el desarrollador que NO comete el bug de truncar "👨‍👩‍👧‍👦" (familia, que son múltiples code points) a la mitad.

**🧠 Feynman Challenge**

> Un colega dice: *"Un string es un array de caracteres"*. ¿Es correcto en Go? Explica la diferencia entre bytes, runes y caracteres. ¿Por qué `len("café")` devuelve 5 y no 4? ¿Qué pasa con emojis como "👨‍👩‍👧‍👦" que en realidad son 7 code points unidos? Si no puedes explicar esto claramente, tu código tiene bugs silenciosos esperando explotar.

</details>

---

#### 📘 [Lección 09 — Interfaces: El Superpoder Oculto de Go](./09-interfaces/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Las interfaces en Go se implementan **implícitamente** — no hay `implements` keyword. Si tu tipo tiene los métodos que la interface requiere, automáticamente implementa esa interface. Esto permite el **duck typing en tiempo de compilación** y desacopla completamente los paquetes.
- **Analogía mental:** En Java, para ser "Swimmable" necesitas firmar un contrato que diga "Yo implemento Swimmable". En Go, si caminas como pato y haces cuac como pato, eres un pato — punto. No necesitas un certificado de "patitud". Esto es el duck typing, pero verificado por el compilador, no en runtime como Python.
- **Caso de uso:** La interface `io.Reader` de Go es implementada por archivos, conexiones de red, buffers, y strings. Esto significa que PUEDES LEER DE CUALQUIER COSA con la misma función — esta es la base de la composición extrema de Go.

**🏋️ Ejercicio Práctico: Sistema de Notificador Multi-Canal**

Construirás un sistema que envía notificaciones por múltiples canales (email, consola, archivo de log, webhook HTTP). Definirás una interface `Notificador` e implementarás múltiples versiones. El sistema seleccionará dinámicamente los canales según configuración.

> **¿Por qué este enfoque?** Porque los sistemas de notificación son un patrón omnipresente en software moderno. Este ejercicio te obliga a diseñar interfaces limpias, implementarlas implícitamente, y demostrar el poder del desacoplamiento: puedes agregar un nuevo canal (Slack, Discord) sin modificar el código existente.

**🧠 Feynman Challenge**

> Explica la diferencia entre interfaces explícitas (Java/C#) e implícitas (Go). ¿Por qué las interfaces implícitas permiten que un código escrito 5 años después implemente tu interface sin modificar tu código original? Usa el ejemplo de `io.Reader`: ¿cómo es posible que un archivo, una conexión de red, y un string usen la misma función `Read()`? Si no puedes responder sin mirar documentación, necesitas repensar interfaces.

</details>

---

### 📙 FASE III — Concurrencia: El Corazón de Go (Lecciones 10–13)
> *Esta es la razón por la que Go fue creado. Aquí te convertirás en un arquitecto de sistemas concurrentes.*

---

#### 📘 [Lección 10 — Goroutines: Miles de Hilos por Centavo](./10-goroutines/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Las goroutines son hilos verdes gestionados por el runtime de Go, no por el OS. Son extremadamente ligeras (~2KB de stack inicial vs ~1MB de un hilo de OS), y el scheduler de Go (basado en el modelo M:N) multiplexa miles de goroutines sobre un número limitado de threads de OS.
- **Analogía mental:** Imagina un restaurante con 10,000 clientes. Un hilo de OS es como un mesero físico — contratar 10,000 meseros es imposible. Una goroutine es como un mesero fantasma ultrarrápido que puede atender miles de mesas simultáneamente, saltando entre ellas cuando un cliente está leyendo el menú (I/O bloqueante).
- **Caso de uso:** Un servidor HTTP en Go puede manejar 100,000+ conexiones concurrentes en una máquina modesta. Node.js usa un event loop, Java usa thread pools — Go usa goroutines, que son más simples que ambos.

**🏋️ Ejercicio Práctico: Scanner de Puertos Concurrente**

Construirás un escáner de puertos TCP que lance miles de goroutines para verificar qué puertos están abiertos en un host dado. Mostrará resultados en tiempo real, manejará timeouts, y respetará un límite de concurrencia configurable (para no saturar la red).

> **¿Por qué este enfoque?** Porque un escáner de puertos es el ejemplo perfecto de I/O-bound concurrency: cada conexión es independiente, la mayoría del tiempo se espera la respuesta de red, y miles de goroutines pueden ejecutarse simultáneamente sin consumir recursos significativos.

**🧠 Feynman Challenge**

> Explica la diferencia entre un hilo del OS y una goroutine usando la analogía del restaurante. ¿Por qué puedes crear 100,000 goroutines pero no 100,000 hilos de OS? ¿Qué es el "GOMAXPROCS" y por qué por defecto es igual al número de CPUs? Si solo dices "las goroutines son más ligeras", no has explicado el *por qué*.

</details>

---

#### 📘 [Lección 11 — Channels: El Sistema Nervioso de Go](./11-channels/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Los channels son el mecanismo de comunicación entre goroutines. Son **tipados**, **bloqueantes** por diseño, y siguen la filosofía de Go: *"Don't communicate by sharing memory; share memory by communicating."* Un channel sincronizado (sin buffer) bloquea al emisor hasta que el receptor está listo — esto es un handshake garantizado.
- **Analogía mental:** Un channel es como una tubería de un banco de drive-through: el cajero (goroutine emisora) pone la cápsula con dinero, y el cliente (goroutine receptora) la recibe. Si la tubería está llena (buffer lleno), el cajero espera. Si está vacía, el cliente espera. No hay pérdida, no hay condición de carrera.
- **Caso de uso:** Los pipelines de procesamiento de datos, fan-out/fan-in, y el patrón worker pool están construidos sobre channels. Es el patrón fundamental de la concurrencia en Go.

**🏋️ Ejercicio Práctico: Pipeline de Descarga y Procesamiento de Archivos**

Construirás un pipeline de 3 etapas usando channels: (1) generador de URLs, (2) downloader concurrente con límite de concurrencia, (3) procesador que calcula hash y extrae metadatos. Cada etapa se comunica exclusivamente por channels.

> **¿Por qué este enfoque?** Porque el patrón pipeline es el "Hello World" de la concurrencia en producción. Este ejercicio te obliga a entender channels buffered vs unbuffered, `select`, y `done channels` para graceful shutdown — los tres conceptos más importantes de la concurrencia en Go.

**🧠 Feynman Challenge**

> Explica la diferencia entre un channel buffered y unbuffered usando la analogía del drive-through. ¿Qué pasa si el emisor envía a un channel lleno? ¿Qué pasa si un receptor lee de un channel vacío? ¿Por qué enviar a un channel `nil` o cerrado causa un panic? Dibuja el diagrama mental de cada escenario.

</details>

---

#### 📘 [Lección 12 — Select: El Director de Orquesta Concurrente](./12-select/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** `select` permite a una goroutine esperar en múltiples channels simultáneamente — como un switch-case, pero para operaciones de channel. Si múltiples channels están listos, `select` elige uno **pseudoaleatoriamente** (no el primero). El caso `default` hace que el select sea no-bloqueante.
- **Analogía mental:** `select` es como un recepcionista de hotel que atiende 5 teléfonos al mismo tiempo. Cuando suena uno, atiende ese. Si suenan 3 a la vez, elige uno (no necesariamente el que sonó primero). Si no suena ninguno, espera (a menos que tenga un `default`, que sería como "hacer otra cosa mientras tanto").
- **Caso de uso:** Timeouts, cancellation, rate limiting, multiplexación de I/O, y cualquier operación que deba responder a múltiples fuentes de eventos.

**🏋️ Ejercicio Práctico: Monitor de Servicios con Timeout y Circuit Breaker**

Construirás un monitor que vigila múltiples servicios HTTP simultáneamente. Usará `select` para manejar timeouts, respuestas exitosas, y errores de cada servicio. Implementará un circuit breaker que deja de monitorear un servicio después de N fallos consecutivos.

> **¿Por qué este enfoque?** Porque el monitoreo de servicios es un problema real de DevOps/SRE. Este ejercicio te obliga a usar `select` con `time.After`, `context.WithTimeout`, y channels de error — los tres patrones más usados en producción con Go.

**🧠 Feynman Challenge**

> Explica por qué `select` elige aleatoriamente entre múltiples channels listos, en vez del primero que esté disponible. ¿Qué problema de starvation se resolvería si eligiera siempre el primero? Luego explica el caso `default` — ¿por qué hace que `select` sea no-bloqueante, y en qué escenario eso es útil vs peligroso (busy loop)?

</details>

---

#### 📘 [Lección 13 — Patrones Avanzados de Concurrencia: Fan-Out, Fan-In, Worker Pools](./13-patrones-concurrencia/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Los patrones de concurrencia son recetas probadas para problemas comunes. **Fan-Out** distribuye trabajo a múltiples goroutines. **Fan-In** converge múltiples channels en uno solo. **Worker Pool** limita el número de goroutines activas. **Semaphore** controla acceso a recursos limitados.
- **Analogía mental:** Fan-Out es como un gerente que reparte tareas entre 10 empleados (goroutines). Fan-In es como un embudo que junta los reportes de los 10 empleados en un solo escritorio (channel). Worker Pool es como un call center con exactamente 5 operadores — si llaman 100 personas, 95 esperan en cola.
- **Caso de uso:** Procesamiento paralelo de imágenes, crawling web, batch processing de bases de datos, y cualquier tarea divisible en unidades independientes.

**🏋️ Ejercicio Práctico: Crawler Web Concurrente con Rate Limiting**

Construirás un web crawler que: (1) usa fan-out para descargar múltiples páginas en paralelo, (2) un worker pool limita las conexiones a N simultáneas, (3) fan-in converge los resultados en un solo stream, (4) un rate limiter respeta los límites de las APIs (token bucket pattern).

> **¿Por qué este enfoque?** Porque un web crawler combina TODOS los patrones de concurrencia en un solo proyecto. Es el ejercicio definitivo de la fase de concurrencia y te prepara para sistemas reales de alta demanda.

**🧠 Feynman Challenge**

> Explica fan-out, fan-in y worker pool a alguien que solo conoce programación secuencial. Usa la analogía de una pizzería: fan-out = repartir pedidos a múltiples cocineros, fan-in = juntar todas las pizzas listas en la caja, worker pool = solo hay 3 cocineros aunque lleguen 50 pedidos. ¿Qué pasa si los cocineros son lentos? ¿Dónde se acumula la cola? ¿Cómo decides cuántos cocineros contratar?

</details>

---

### 📕 FASE IV — Ingeniería y Arquitectura (Lecciones 14–17)
> *De programador a ingeniero. Aquí aprenderás a construir sistemas profesionales.*

---

#### 📘 [Lección 14 — Paquetes, Módulos y la Organización del Código Go](./14-paquetes-modulos/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go organiza el código en **paquetes** (directorios) y gestiona dependencias con **Go Modules** (`go.mod`, `go.sum`). Las reglas de visibilidad son elegantes: si empieza con mayúscula es exportado, si empieza con minúscula es privado. No hay `public`, `private`, `protected` — solo capitalización.
- **Analogía mental:** Un paquete de Go es como una carpeta de herramientas con una etiqueta clara: las herramientas pintadas de rojo (mayúscula) pueden ser prestadas a otros talleres, las pintadas de gris (minúscula) son solo para uso interno. Los `go modules` son como el inventario centralizado que registra qué herramientas tienes y de quién las pediste prestadas.
- **Caso de uso:** Todo proyecto Go profesional usa modules. Entender `go mod init`, `go mod tidy`, `go get`, y los replace directives es absolutamente esencial.

**🏋️ Ejercicio Práctico: Librería de Utilidades Reutilizable**

Construirás tu propia librería de utilidades (`myutils`) organizada en subpaquetes: `myutils/math` (funciones matemáticas), `myutils/text` (procesamiento de texto), `myutils/file` (operaciones de archivo). Publicarás funciones exportadas y privadas, la documentarás con godoc, y la usarás desde un programa principal.

> **¿Por qué este enfoque?** Porque crear una librería reutilizable te enseña las reglas de visibilidad, la organización de paquetes, los ciclos de importación (prohibidos en Go), y cómo escribir código documentado — habilidades que todo desarrollador Go profesional necesita.

**🧠 Feynman Challenge**

> Explica por qué Go eligió la capitalización como mecanismo de visibilidad en vez de keywords como `public`/`private`. ¿Qué ventajas tiene? ¿Qué desventajas? Luego explica qué es un ciclo de importación y por qué Go lo prohíbe. Si tu respuesta es "porque sí", no has entendido la arquitectura de paquetes.

</details>

---

#### 📘 [Lección 15 — Testing en Go: El Estándar de Oro](./15-testing/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go tiene testing integrado en la librería estándar — no necesitas frameworks externos. Los archivos `_test.go` se excluyen del build normal. `go test` ejecuta tests, benchmarks, y examples. Las **table-driven tests** son el patrón estándar de la industria: defines una tabla de casos de prueba y los ejecutas en un bucle.
- **Analogía mental:** Testing en Go es como un inspector de calidad en una fábrica de automóviles: no necesitas contratar a una empresa externa (framework de testing), el inspector viene incluido con la fábrica. Las table-driven tests son como una lista de chequeo que inspecciona el mismo componente en 50 configuraciones diferentes.
- **Caso de uso:** En Go, se espera que TODO código profesional tenga tests. Las empresas de Silicon Valley rechazan PRs sin tests, y `go test -cover` te dice exactamente qué porcentaje de tu código está cubierto.

**🏋️ Ejercicio Práctico: Suite de Tests para Tu Librería de Utilidades**

Escribirás tests comprehensivos para la librería de la Lección 14 usando table-driven tests, subtests, tests de benchmark (`BenchmarkXxx`), y tests de ejemplo (`ExampleXxx` que también sirven como documentación). Medirás la cobertura y la mejorarás al 90%+.

> **¿Por qué este enfoque?** Porque testear tu propio código (no código ajeno) te obliga a pensar como un adversario: ¿qué pasa con inputs vacíos? ¿con números negativos? ¿con strings de 1 millón de caracteres? Este ejercicio forja el mindset de un ingeniero de software profesional.

**🧠 Feynman Challenge**

> Explica qué son las table-driven tests y por qué son mejores que escribir tests individuales para cada caso. Luego explica la diferencia entre un test, un benchmark y un example en Go. ¿Por qué los examples son también tests? ¿Y por qué los benchmarks son cruciales para código de alto rendimiento? Si solo dices "son para medir velocidad", profundiza más.

</details>

---

#### 📘 [Lección 16 — Paquete `os`, `io` y el Mundo de los Archivos y Procesos](./16-os-io/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go brilla en programación de sistemas. Los paquetes `os` e `io` proporcionan abstracciones elegantes para archivos, directorios, variables de entorno, y procesos del OS. La interface `io.Reader`/`io.Writer` es el pilar fundamental: TODO en Go que produce o consume datos implementa una de estas dos interfaces.
- **Analogía mental:** `io.Reader` y `io.Writer` son los enchufes universales de Go. Un archivo es una fuente de datos, una conexión de red es una fuente de datos, un buffer en memoria es una fuente de datos — todas implementan `Reader`. Esto significa que puedes "enchufar" cualquier fuente a cualquier destino sin cables especiales.
- **Caso de uso:** Herramientas CLI que procesan archivos, servidores de archivos, proxies de red, y pipelines de datos usan estos paquetes extensivamente.

**🏋️ Ejercicio Práctico: Herramienta CLI de Backup Incremental**

Construirás una herramienta CLI que: (1) escanea un directorio recursivamente, (2) identifica archivos modificados desde el último backup (comparando timestamps), (3) los copia a un directorio de destino preservando la estructura, (4) genera un archivo de log con los cambios, y (5) acepta flags desde la línea de comandos con el paquete `flag`.

> **¿Por qué este enfoque?** Porque un backup incremental es una herramienta real y útil que usa `os` para archivos y directorios, `io` para copiar datos eficientemente, `filepath` para manipular paths, y `flag` para la CLI. Es un proyecto completo que consolida todo lo aprendido.

**🧠 Feynman Challenge**

> Explica por qué `io.Reader` y `io.Writer` son las interfaces más importantes de Go. ¿Qué significa que "todo es un stream"? ¿Por qué puedes pasar el resultado de un archivo directamente a una conexión HTTP con `io.Copy` sin leer todo el archivo a memoria? ¿Qué es un `io.MultiWriter` y cuándo lo usarías?

</details>

---

#### 📘 [Lección 17 — JSON, Serialización y APIs REST](./17-json-apis/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go tiene un marshaling JSON basado en **struct tags** — metadatos en los campos del struct que controlan cómo se serializa/deserializa. El paquete `encoding/json` usa reflection internamente, pero la API es simple: `json.Marshal`, `json.Unmarshal`, `json.NewEncoder`, `json.NewDecoder`.
- **Analogía mental:** Los struct tags son como etiquetas adhesivas en una mudanza: le dicen a los movers (el marshaller) "esta caja se llama 'nombre' en el nuevo apartamento" (`json:"nombre"`). Sin etiquetas, los movers usan el nombre original de la caja (el nombre del campo en Go).
- **Caso de uso:** Cada API REST del planeta habla JSON. Entender cómo serializar/deserializar JSON en Go es tan esencial como saber respirar para un backend developer.

**🏋️ Ejercicio Práctico: Cliente de API REST con Cache Local**

Construirás un CLI que consume una API pública real (GitHub API, OpenWeatherMap, o PokeAPI), deserializa la respuesta JSON en structs tipados, implementa cache local en archivo JSON para evitar requests repetidos, y muestra los datos formateados en una tabla en consola.

> **¿Por qué este enfoque?** Porque consumir APIs REST es la tarea #1 de un backend developer. Este ejercicio te obliga a trabajar con JSON anidado, optional fields, custom unmarshaling, y el diseño de structs que reflejen la estructura de la API.

**🧠 Feynman Challenge**

> Explica qué son los struct tags en Go y por qué `json:"nombre,omitempty"` es diferente de `json:"nombre"`. ¿Qué hace `omitempty`? ¿Qué pasa si un campo del JSON no tiene un campo correspondiente en el struct? ¿Y si el struct tiene un campo que no existe en el JSON? Explica cada escenario sin mirar documentación.

</details>

---

### 📓 FASE V — Nivel Experto (Lecciones 18–20)
> *Domina los temas que separan al senior del junior. Aquí forjas maestría.*

---

#### 📘 [Lección 18 — `context`, `defer`, `panic/recover` y el Control de Flujo Avanzado](./18-context-defer/)

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** `context` es el mecanismo de Go para propagar cancellation, timeouts y valores de request a través de goroutines y llamadas a servicios. `defer` garantiza la limpieza de recursos (LIFO stack). `panic/recover` es el mecanismo de emergencia para situaciones que no deberían ocurrir — NO es un sistema de excepciones.
- **Analogía mental:** `context` es como un pasaporte que viaja contigo a través de todo el sistema: lleva tu identidad (valores), tu tiempo límite (deadline), y si te cancelan el vuelo (cancel), TODAS las goroutines que usan ese pasaporte se enteran inmediatamente y dejan de trabajar.
- **Caso de uso:** Todo servidor HTTP profesional en Go pasa `context.Context` como primer parámetro. Los orígenes de los microservicios de Netflix, Uber, y Cloudflare están llenos de `ctx context.Context`.

**🏋️ Ejercicio Práctico: Proxy HTTP con Timeout y Rate Limiting por IP**

Construirás un proxy HTTP que: (1) propaga context con timeout a los servicios backend, (2) usa `defer` para logging y cleanup de conexiones, (3) implementa rate limiting por IP usando context values, (4) usa `recover` para capturar panics en goroutines sin tumbar el servidor.

> **¿Por qué este enfoque?** Porque un proxy HTTP es el escenario perfecto donde `context`, `defer` y `panic/recover` trabajan juntos. Este ejercicio te expone al código que escribe un Staff Engineer en producción.

**🧠 Feynman Challenge**

> Explica `context.Context` sin usar la palabra "context". ¿Qué es realmente? ¿Un timeout? ¿Un cancelador? ¿Un diccionario? Explica por qué `context.Background()` es el root, por qué `context.TODO()` existe, y por qué nunca deberías guardar un context en un struct. Si dices "porque el linter te lo dice", no has entendido el *por qué*.

</details>

---

#### 📘 [Lección 19 — Generics, Reflect y Metaprogramación en Go](./19-generics-reflect/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Go 1.18 introdujo **generics** (type parameters) — la feature más solicitada en la historia del lenguaje. Los generics permiten escribir funciones y tipos que trabajan con cualquier tipo que satisfaga constraints. El paquete `reflect` permite inspeccionar tipos en runtime — es poderoso pero peligroso y lento.
- **Analogía mental:** Sin generics, cada herramienta solo funciona con un material: tienes un martillo para clavos de madera y OTRO martillo para clavos de metal. Con generics, tienes un martillo que acepta clavos de CUALQUIER material siempre y cuando sean clavos (satisfagan el constraint). `reflect` es el martillo que acepta cualquier cosa, incluso cosas que no son clavos — peligroso pero a veces necesario.
- **Caso de uso:** Las colecciones genéricas (slices, maps), las funciones utilitarias (`Map`, `Filter`, `Reduce`), y las bibliotecas de serialización usan generics y reflect.

**🏋️ Ejercicio Práctico: Librería Genérica de Colecciones Funcionales**

Construirás una librería genérica que implemente `Map`, `Filter`, `Reduce`, `Contains`, `Unique`, y `Sort` para slices de cualquier tipo. Luego implementarás un validador de structs usando `reflect` que lea tags personalizados (`validate:"required,min=3,max=50"`) y valide automáticamente cualquier struct.

> **¿Por qué este enfoque?** Porque una librería funcional genérica es el ejercicio definitivo para entender type parameters, constraints, y la diferencia entre `any` y constraints específicos. El validador de structs con `reflect` te enseña metaprogramación real — el mismo patrón usado por librerías como `go-playground/validator`.

**🧠 Feynman Challenge**

> Explica la diferencia entre `any`, `comparable`, y un constraint personalizado como `[T Number]`. ¿Por qué `[T any]` no te deja comparar valores con `==`? ¿Qué es un constraint interface y cómo difiere de una interface normal? Luego explica por qué `reflect` es 10-100x más lento que código genérico — si no sabes por qué, no entiendes el costo de la reflexión en runtime.

</details>

---

#### 📘 [Lección 20 — Proyecto Final: Construye una Base de Datos KV con Persistencia](./20-proyecto-final/) ✅

<details>
<summary><strong>🔍 Expandir detalles de la lección</strong></summary>

<br>

**🎯 Enfoque de la Lección**

- **Objetivo conceptual:** Esta lección integra TODO lo aprendido en un proyecto ambicioso: una base de datos key-value con persistencia en disco, transacciones simples, y un servidor TCP que acepta comandos estilo Redis. Es el equivalente a tu "examen final" donde cada concepto se aplica en su justa medida.
- **Analogía mental:** Construir esta base de datos es como construir un reloj suizo: cada engranaje (goroutines, channels, structs, interfaces, archivos, JSON) debe funcionar en perfecta sincronía. Si un engranaje falla, todo el reloj se detiene. Es la prueba definitiva de tu comprensión integral.
- **Caso de uso:** Redis, BoltDB, LevelDB y PebbleDB son bases de datos que empezaron como proyectos similares. Entender cómo funcionan por dentro te da superpoderes para debuggear problemas de rendimiento, entender trade-offs de almacenamiento, y diseñar sistemas de datos.

**🏋️ Ejercicio Proyecto Final: GoKV — Base de Datos Key-Value**

Construirás **GoKV**, una mini base de datos con:

| Componente | Conceptos de Go Aplicados |
|:-----------|:--------------------------|
| Motor de almacenamiento | Structs, maps, interfaces, `io.Reader`/`io.Writer` |
| Persistencia en disco | `os` package, `encoding/binary`, WAL (Write-Ahead Log) |
| Servidor TCP | Goroutines, `net` package, `bufio.Scanner` |
| Protocolo de comandos | String parsing, `strings.Fields`, switch |
| Transacciones | `sync.RWMutex`, atomic operations |
| CLI cliente | `flag` package, `net.Conn`, formatting |
| Tests comprehensivos | Table-driven tests, benchmarks, examples |
| Documentación | godoc, README |

**Comandos soportados:**
```
SET key value    → Almacena un par clave-valor
GET key          → Recupera el valor de una clave
DEL key          → Elimina una clave
BEGIN            → Inicia una transacción
COMMIT           → Confirma la transacción
ROLLBACK         → Revierte la transacción
STATS            → Muestra estadísticas del servidor
```

**🧠 Feynman Challenge**

> Presenta tu proyecto GoKV a alguien (o grábate explicándolo). Debes ser capaz de explicar: (1) por qué usaste un WAL en vez de serializar todo el mapa a disco, (2) cómo los locks previenen corrupción de datos, (3) por qué cada conexión TCP es una goroutine, (4) qué pasaría si dos clientes escriben la misma clave simultáneamente, y (5) cómo harías para escalar esto a 1 millón de conexiones. Si puedes explicar los 5 puntos, has dominado Go.

</details>

---

## 📊 Resumen por Fases

```
┌─────────────────────────────────────────────────────────────────┐
│                                                                 │
│   📗 FASE I: Fundamentos          ██░░░░░░░░░░░░░░░░░░  25%   │
│      Lecciones 01-05                                             │
│      Variables · Tipos · Control · Funciones · Structs          │
│                                                                 │
│   📘 FASE II: Tipos de Datos      ████████░░░░░░░░░░░░  45%   │
│      Lecciones 06-09                                             │
│      Slices · Maps · Strings · Interfaces                       │
│                                                                 │
│   📙 FASE III: Concurrencia       ████████████░░░░░░░░  65%   │
│      Lecciones 10-13                                             │
│      Goroutines · Channels · Select · Patrones                  │
│                                                                 │
│   📕 FASE IV: Ingeniería          ████████████████░░░░  85%   │
│      Lecciones 14-17                                             │
│      Paquetes · Testing · OS/IO · JSON/APIs                     │
│                                                                 │
│   📓 FASE V: Nivel Experto        ████████████████████  100%  │
│      Lecciones 18-20                                             │
│      Context · Generics · Proyecto Final                        │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## ⚡ Instrucciones de Uso

1. **Sigue el orden.** Cada lección construye sobre la anterior
2. **Lee la lección completa** antes de mirar las soluciones
3. **Intenta el ejercicio** antes del Feynman Challenge
4. **Habla en voz alta** durante el Feynman Challenge — escribir no es suficiente
5. **Si no puedes explicarlo simple**, vuelve a la lección
6. **Crea un directorio** por cada lección (`01-hello-go/`, `02-variables/`, etc.)

```bash
# Estructura del repositorio
go-curso/
├── 00-indice.md          ← Estás aquí
├── 01-hello-go/
│   ├── main.go
│   └── go.mod
├── 02-variables/
│   ├── main.go
│   └── go.mod
├── 03-control-flujo/
│   ├── main.go
│   └── go.mod
├── ...
└── 20-proyecto-final/
    ├── 20-proyecto-final.md
    ├── main.go
    └── go.mod
```

---

<div align="center">

### *"La única forma de aprender es construir cosas."*
### — **Rob Pike**, co-creador de Go

<br>

**¡Comienza con la Lección 01 y que arranque el laboratorio! 🧪**

</div>