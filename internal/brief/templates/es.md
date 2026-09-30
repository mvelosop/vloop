---
name: {{NAME}}
description: <una línea, empezando por verbo — qué existe tras la ejecución que hoy no existe>
kind: brief
status: draft
created: {{CREATED}}
seeds: Una ejecución que <construye qué, en qué orden>
depends-on: []
---
# Brief — <una línea: qué construye esto>

- **Punto de partida:** desde cero / extiende `<rama>` en `<sha>`. El planificador
  fija ese SHA como la **base** con la que se comparan todas las puertas — nunca
  `HEAD~n`.
- **Origen:** <el acto de diseño o el brief de arquitectura del que sale esto, o
  "decisión del operador, AAAA-MM-DD">

> Sustituye todo lo que hay debajo del frontmatter. El nombre del archivo es
> `docs/briefs/B<AAAAMMDD-HHMM>-<slug>.loop-brief.md`: numerado por la hora de
> creación, al minuto, porque un contador colisiona en cuanto dos personas
> redactan en ramas distintas. El sufijo `.loop-brief` nombra al consumidor; el
> id de la ejecución, y con él el nombre del diario, es el nombre del archivo sin
> ese sufijo.
>
> El `status` del frontmatter es la única marca de planificación: `draft` mientras
> lo escribes, `ready` cuando está listo para planificar, `consumed` una vez que
> una ejecución lo ha usado, `abandoned` si lo descartas. Enumera en `depends-on`, por nombre, los briefs sobre
> los que se apoya este. Pasa `vloop brief check` antes de gastar nada en él.
>
> La regla que decide si un brief es bueno: **fija las decisiones, deja abierta la
> mecánica.** Nombra el comportamiento, los formatos, los casos de error, el
> ejemplo trabajado. No nombres la estructura de módulos, las clases ni la
> estructura de archivos: eso le toca a quien lo implemente, y fijarlo no aporta
> nada y sí limita el plan.

---

## Qué es

<Dos o tres frases. Qué existe al final que hoy no existe, y para quién. Si no
puedes decirlo en tres frases, el brief abarca demasiado; divídelo.>

<En un brief de corrección, explica por qué se colaron los defectos: qué punto
ciego comparten las puertas. Una corrección que no cierra ese punto ciego reabre
el defecto la próxima vez que alguien toque el archivo.>

## Por qué esta forma, y qué se descartó

Decidido con el operador el <AAAA-MM-DD>. **No se vuelve a discutir**: el
planificador lo hereda.

- **<La decisión.>** <Una o dos frases de por qué.>
  *Descartado:* <la alternativa que alguien propondrá, y por qué perdió>.
- **<La siguiente decisión.>** <Por qué.>

<Esta sección evita que un planificador, una sesión de trabajo o un revisor
reabra una bifurcación ya resuelta a mitad de la ejecución: cada uno es una
sesión nueva que no vio la discusión. Una opción descartada y escrita es una que
nadie gasta un intento en redescubrir.>

## Referencias vinculantes

- `<ruta relativa al repo>` — <por qué vincula: qué restringe en esta ejecución>

<Cada entrada es una ruta entre comillas invertidas, y luego un motivo. Una ruta
que no resuelve, o un motivo que falta, es un problema que la comprobación
señala. Esta sección es la única forma en que un documento vincula una tarea.>

## Contrato de comportamiento

<Las decisiones. Sé exacto en todo lo que una puerta vaya a afirmar: códigos de
salida, códigos de estado, formatos de salida, mensajes de error, qué es
idempotente, qué persiste, orden, determinismo.>

<Cuando un requisito tenga una consecuencia de diseño, dilo. "Esto debe poder
probarse" es una restricción sobre la implementación, y dejarla implícita es como
se acaba con un comportamiento que nadie puede comprobar.>

<O bien escribe aquí el contrato, o bien delégalo ENTERO en un único documento
nombrado: "el comportamiento es exactamente lo que especifica `docs/<spec>.md`, y
nada más". Nunca mitad aquí y mitad allí: se desincronizan y nada desempata.>

### Decidido aquí, porque los documentos citados lo dejan indeterminado

- <Un hueco en la especificación que el planificador rellenaría por accidente, y
  cómo se rellena.>

### Violaciones sobre las que la revisión debe pronunciarse

- <Una forma de error que pasa la puerta pero debe suspender la revisión: una
  regresión que las pruebas no ven, una desviación de una decisión citada,
  alcance tomado de una porción hermana.>

<En un brief de corrección, una subsección por defecto, cada una en tres partes:>

### F1 — <qué puede hacer ahora el usuario>

**Defecto.** <Qué ocurre hoy, dónde, y el mecanismo.>

**Requerido.** <El comportamiento tras la corrección. Fija el resultado, no el
código.>

**Puerta.** <La afirmación que se entrega junto a la corrección. Debe fallar
contra la implementación actual: demuéstralo antes de corregir.>

## Ejemplo trabajado

<La prueba de aceptación de punta a punta, con valores concretos. Es lo que
arbitra cuando dos implementaciones discrepan, así que debe ser exacto, no
ilustrativo.>

```
<entrada>
  -> <salida esperada exacta>                                  salida 0

<el caso de error>
  -> <fallo esperado exacto>                                   salida 1
```

<Incluye los casos de fallo que una comprobación debe detectar, plantados en una
copia de trabajo o en un fixture: una comprobación que nunca se ha visto fallar
no demuestra nada.>

<Si un valor se genera o se cuenta a partir del árbol en lugar de ser fijo, di
que la puerta debe usar el valor que calcula en vez de codificar uno. Un conteo
codificado es correcto el día que se escribe el plan e incorrecto el día que
entra cualquier otra cosa.>

## Fuera de alcance

<No versiones más pequeñas de esto. Nada de esto. Esta lista es lo que convierte
el desvío de alcance en un hallazgo y no en una cuestión de gusto, así que nombra
lo que alguien añadiría con verosimilitud. Cuando el diseño se cortó en porciones
ordenadas, las porciones hermanas SON esta lista.>

- <la función adyacente más obvia>
- <la deriva que esta ejecución dejará al descubierto pero no debe corregir: la
  informa>
- <lo que un generador de andamiaje crearía sin que nadie lo pida>
- <la mejora "ya que estamos">

## Restricciones

- <lenguaje, entorno de ejecución, gestor de paquetes, plataforma — p. ej. debe
  funcionar con las utilidades BSD de macOS>
- <qué no puede usarse: red en tiempo de ejecución, servicios externos,
  dependencias extra>
- <qué tan rápidas deben ser las puertas: se repiten en cada iteración>
- **Las puertas siguen la superficie del cambio.** Cada tarea se comprueba contra
  lo que toca. Ninguna tarea ejecuta toda la batería de pruebas: una prueba
  inestable en código que la rama nunca tocó bloqueará una tarea correcta,
  intento tras intento.
- **Ninguna tarea puede debilitar una comprobación para pasar.** Una comprobación
  errónea es una tarea bloqueada con una nota, no una edición de la comprobación.
- **Las herramientas nuevas se prueban con fixtures**: nombra los casos que cada
  fixture debe cubrir, incluidos aquellos en los que debe fallar.
- Rutas relativas al repo en todas partes. Ninguna ruta absoluta en archivos ni
  en mensajes de commit.

## Forma

De <N> a <M> tareas, cada una verificable de forma independiente con un solo
comando. El orden es determinante:

1. **<La comprobación o herramienta en la que se apoyan todas las demás>, con sus
   fixtures.** Se comprueba contra los fixtures. Sobre el árbol real puede que
   todavía no esté en verde.
2. **<La siguiente porción.>** Se comprueba con <el comando>.
3. **Cierre:** todas las puertas a la vez, y el ejemplo trabajado línea por línea.

<Si algún comportamiento necesita su propia puerta, dilo aquí. Un artefacto de
larga vida que debe arrancar, una migración que debe aplicarse, un camino de
punta a punta que ninguna prueba unitaria cubre: nómbralo, o acabará dentro de
una tarea que no lo demuestra.>

<Comprueba la tarea de andamiaje contra algo que el proyecto produzca, no contra
una ejecución de pruebas vacía. Un ejecutor de pruebas sin pruebas no siempre
sale con 0.>

<Si una tarea es sobre todo prosa y ningún script puede comprobarla, di qué pasa
cuando suspende la revisión dos veces por redacción y no por una puerta:
bloquearla, y el operador la termina a mano.>

<!--
Tras la ejecución, en este archivo:

- Pon `status: consumed` en el frontmatter.
- Añade una sección `## Registro de la ejecución`: resultado, ejecuciones y gasto,
  qué cambió el operador a mano, y qué sigue abierto. El diario es la vista por
  iteración; esto es el resumen de una pantalla de toda la ejecución.
-->
