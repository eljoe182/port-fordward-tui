# PRD — Kubernetes Port-Forward TUI

Fecha: 2026-04-16
Estado: Draft validado con el usuario

## 1. Executive Summary

### Problem Statement
El flujo actual para levantar port-forwards depende de un script shell estático que obliga a editar código cuando aparecen nuevos targets y genera fricción al cambiar contexto, namespace, selección y puertos locales. Esto vuelve lento un uso frecuente y operativo.

### Proposed Solution
Construir una TUI en Go + Bubble Tea que replique el comportamiento base del script actual como MVP y lo evolucione a un workspace ligero para descubrir targets dinámicamente, configurar favoritos/recientes/puertos preferidos y operar forwards activos sin reiniciar la sesión.

### Success Criteria
- El usuario puede agregar nuevos targets sin editar el script ni tocar código manualmente.
- El tiempo para levantar forwards frecuentes se reduce perceptiblemente mediante favoritos, recientes y puertos predefinidos editables.
- La app reduce errores operativos asociados a colisiones de puertos, selección repetitiva y cambios manuales.
- La TUI permite iniciar, detener y reintentar forwards individuales sin tumbar los demás.
- La app se distribuye como binario simple para macOS/Linux y funciona con `kubectl` disponible en el sistema.

## 2. User Experience & Functionality

### User Personas
- **Developer/operator principal**: usa port-forward múltiples veces al día, cambia contexto/namespace/servicios con frecuencia y necesita velocidad operativa.
- **Usuario técnico secundario**: necesita levantar forwards de forma ocasional, pero valora una UX clara y menos dependiente de scripts manuales.

### User Stories

#### Story 1 — Selección de contexto y namespace
Como usuario, quiero seleccionar contexto y namespace desde la TUI para no depender de comandos manuales separados.

**Acceptance Criteria**
- La app muestra el contexto actual de `kubectl`.
- La app lista contextos disponibles y permite cambiar el contexto activo.
- La app lista namespaces del contexto activo y permite elegir uno.
- La app permite refrescar descubrimiento después de cambiar contexto o namespace.

#### Story 2 — Descubrimiento híbrido de targets
Como usuario, quiero ver pods y services disponibles sin editar scripts para agregar nuevos targets.

**Acceptance Criteria**
- La app descubre dinámicamente targets del namespace actual.
- La app soporta targets de tipo `service` y `pod`.
- La app mezcla targets descubiertos con targets configurados por el usuario.
- Si un target existe tanto descubierto como configurado, se muestra como una sola entidad enriquecida.
- La vista principal prioriza recursos útiles y permite cambiar criterio de orden/filtro.

#### Story 3 — Configuración operativa reutilizable
Como usuario, quiero guardar favoritos, alias y puertos preferidos para acelerar sesiones futuras.

**Acceptance Criteria**
- La app persiste favoritos entre ejecuciones.
- La app persiste recientes entre ejecuciones.
- La app permite definir alias amigables por target.
- La app persiste puerto local preferido por target cuando aplique.
- La app puede preferir `pod` o `service` según configuración previa cuando ambos sean válidos.

#### Story 4 — Selección múltiple y edición de puertos
Como usuario, quiero seleccionar múltiples targets y editar sus puertos locales antes de iniciar para mantener defaults útiles pero con override rápido.

**Acceptance Criteria**
- La app permite selección múltiple desde el catálogo principal.
- Cada target muestra puerto remoto y puerto local sugerido.
- El puerto local sugerido puede editarse antes de iniciar el forward.
- La app valida colisiones dentro de la selección actual.
- La app advierte o bloquea cuando un puerto ya está ocupado.

#### Story 5 — Operación continua de forwards
Como usuario, quiero mantener algunos forwards activos y agregar o cambiar otros sin reiniciar la sesión completa.

**Acceptance Criteria**
- La app permite iniciar nuevos forwards sin detener los ya activos.
- La app muestra estados por forward: `starting`, `active`, `failed`, `stopped`.
- La app permite detener forwards individuales.
- La app permite reintentar forwards fallidos.
- La app muestra errores breves y accionables por target.

#### Story 6 — Workspace ligero orientado a operación
Como usuario, quiero una interfaz de trabajo estable y visible para operar forwards de forma continua, no un wizard desechable.

**Acceptance Criteria**
- La TUI usa layout de workspace ligero.
- La vista principal incluye catálogo central de targets.
- La interfaz incluye panel derecho con tabs `Selected` y `Running`.
- La tab `Selected` permite revisar selección y editar puertos.
- La tab `Running` muestra activos/fallidos con acciones `stop` y `retry`.

### Non-Goals
- No incluir health checks avanzados en V1.
- No incluir restauración completa de sesión en V1.
- No incluir presets complejos o stacks automáticos en V1.
- No incluir observabilidad profunda o dashboard avanzado en V1.
- No reemplazar `kubectl`; la app se apoya en él.

## 3. AI System Requirements (If Applicable)

No aplica. Este producto no depende de capacidades de IA para su funcionamiento principal.

## 4. Technical Specifications

### Architecture Overview
La solución se implementará como una aplicación en Go con Bubble Tea para la capa TUI. La lógica debe separarse en componentes claros:

- **Kubernetes Discovery**: obtiene contextos, namespace y targets forwardeables.
- **Catalog Resolver**: mezcla descubrimiento dinámico con configuración persistida y calcula orden `Smart`.
- **Forward Runtime**: inicia y supervisa procesos `kubectl port-forward`, administra ciclo de vida y estados.
- **Persistence**: guarda favoritos, recientes, alias, puertos preferidos y preferencias mínimas.
- **TUI State Layer**: administra navegación, selección, tabs, edición de puertos y renderizado de estados.

### Functional Model
- **Target**
  - `id`
  - `type` = `service | pod`
  - `name`
  - `namespace`
  - `alias?`
  - `remotePort`
  - `preferredLocalPort?`
  - `source` = `discovered | configured | merged`
  - `favorite`
  - `recentMetadata?`

- **ForwardSession**
  - `targetId`
  - `localPort`
  - `remotePort`
  - `status`
  - `processHandle`
  - `startedAt`
  - `lastError?`

- **AppConfig**
  - favoritos
  - recientes
  - aliases
  - puertos preferidos
  - preferencias mínimas de orden/filtro

### Discovery & Ranking Rules
- La app debe soportar descubrimiento híbrido configurable.
- El catálogo principal usa orden por defecto `Smart`.
- El ranking `Smart` se apoya en señales visibles:
  - favorito
  - recencia
  - configuración guardada
  - relevancia en namespace actual
  - estado usable del recurso
  - coincidencia con búsqueda
  - opcionalmente si ya está corriendo
- La app debe ofrecer órdenes alternativos: `Name`, `Recent`, `Favorites`, `Type`.

### Runtime Behavior
- Debe ser no destructivo: agregar nuevos forwards no debe derribar los existentes.
- Debe soportar cleanup ordenado al salir.
- Debe capturar errores de `kubectl` y traducirlos a mensajes breves y accionables.

### Integration Points
- **CLI del sistema**: `kubectl`
- **Kubernetes context config**: configuración local usada por `kubectl`
- **Persistencia local**: archivo(s) locales de configuración y estado liviano

### Security & Privacy
- La app no debe almacenar secretos sensibles del cluster.
- Debe reutilizar la autenticación ya resuelta por `kubectl`.
- La persistencia local debe limitarse a preferencias operativas y metadata no sensible.

### Platform Constraints
- V1 debe soportar **macOS y Linux**.
- V1 debe funcionar donde funcione `kubectl`, sin dependencias exóticas adicionales.
- V1 debe distribuirse como **binario simple**.

### Testing Expectations
- Tests unitarios para merge de targets configurados + descubiertos.
- Tests unitarios para ranking `Smart`.
- Tests unitarios para validación de puertos y conflictos.
- Tests de lógica para transiciones de estado del runtime.
- Tests para persistencia local de favoritos/recientes/configuración.
- Tests de TUI acotados a navegación y estados críticos, sin acoplar toda la lógica a renderizado.

## 5. Risks & Roadmap

### Phased Rollout

#### MVP / V1
- Paridad funcional con el script actual:
  - elegir contexto
  - elegir namespace
  - seleccionar targets
  - iniciar múltiples port-forwards
  - cleanup al salir
- Mejoras clave:
  - descubrimiento híbrido
  - favoritos/recientes persistentes
  - puertos sugeridos editables
  - workspace ligero
  - runtime con `stop` y `retry`

#### V1.1
- Refinamiento de ranking `Smart`
- Mejoras de filtros y búsqueda
- Mejores mensajes de error y sugerencias
- UX polishing del panel `Selected` / `Running`

#### V2.0 Pro
- perfiles por proyecto/contexto/namespace
- presets o grupos de forwards
- restauración de sesión
- health checks por target
- observabilidad/logs más ricos
- automatizaciones operativas avanzadas

### Technical Risks
- Descubrimiento demasiado ruidoso si se muestran demasiados recursos por defecto.
- Mala unificación entre `pod` y `service`, afectando la UX.
- Acoplar demasiado Bubble Tea con procesos reales y volver frágil el runtime.
- Ranking `Smart` percibido como arbitrario si no se explica o controla.
- Manejo inestable del ciclo de vida de procesos `kubectl port-forward`.

### Mitigations
- Priorizar targets útiles primero y permitir filtros/órdenes explícitos.
- Modelar `Target` como entidad unificada con tipo explícito.
- Separar runtime de procesos de la capa de UI.
- Hacer visible el orden `Smart` y permitir órdenes alternativos.
- Modelar estados explícitos y cleanup robusto.
