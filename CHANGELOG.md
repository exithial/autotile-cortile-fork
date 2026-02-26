# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.10.0] - 2026-02-26

### Added

- **Lógica de master inteligente en autotile**: El comando `Ctrl+Shift+k5` ahora establece el master según número de columnas: 2→derecha, 3→centro, 4→centro-derecha
- **Distribución priorizada**: La columna master es la última en dividirse verticalmente, priorizando divisiones en columnas slave

### Fixed

- **Navegación con k2**: Corregido bug donde `Ctrl+Shift+k2` (next window) se ejecutaba dos veces debido a línea duplicada
- **Algoritmo de distribución**: Ventanas extra se distribuyen primero en slaves hasta que tengan ≥2 ventanas, luego en master

### Technical Details

- `layout/autotile.go`: Nuevo algoritmo en `applyColumns()` que prioriza slaves. Métodos `determineMasterColumn()`, `determineMasterIndex()` actualizados
- `input/action.go`: Eliminada línea duplicada en `case "window_next"`
- **Lógica master**: Para N ventanas y C columnas, master se calcula dinámicamente según distribución actual

## [2.9.0] - 2026-02-26

### Added

- **Redimensión automática de ventanas flotantes**: Las ventanas flotantes se redimensionan automáticamente: 1920x1080 en pantallas ultrawide (>2560px), 80% del escritorio en pantallas estándar
- **Exclusión de ventana plasmashell**: Se excluye la interfaz de KDE (`plasmashell`) del tiling para evitar conflictos con Plasma

### Technical Details

- `desktop/tracker.go`: Método `ToggleFloat(w)` ahora calcula dimensiones según `ultrawide_threshold`. Si `dw > ultrawide_threshold` → 1920x1080, sino → 80% del escritorio. Centra la ventana con `x = dx + (dw-newW)/2`, `y = dy + (dh-newH)/2`
- `config.toml`: Agregado `["plasmashell", ""]` en `window_ignore`
## [2.8.0] - 2026-02-25

### Added

- **Toggle de ventana flotante**: Nuevo atajo `Control-Shift-G` que extrae la ventana activa del tiling y la deja flotante. Al presionar de nuevo, la ventana vuelve al layout activo
- **Feedback visual del toggle flotante**: Al activar/desactivar el toggle se muestra un overlay representativo — un diagrama de ventana flotante sobre fondo tileado (`floating`) o el layout actual con todos los clientes (`tiling`)

### Technical Details

- `desktop/tracker.go`: Campo `FloatedWindows map[xproto.Window]bool` en `Tracker`. Método `ToggleFloat(w)` que agrega/elimina del mapa y fuerza `Update()` + `Tile()`. `isTrackable()` consulta el mapa para excluir ventanas flotantes sin depender de estados EWMH
- `store/client.go`: Función helper `IsAbove(info)` para detectar `_NET_WM_STATE_ABOVE`
- `input/action.go`: Nuevo `case "window_float_toggle"` en `ExecuteAction()`. Función `ToggleWindowFloat(tr)` que obtiene la ventana activa via `store.Windows.Active.Id` (funcional aunque no esté trackeada)
- `ui/overlay.go`: Función `ShowWindowFloat(ws, floated)` con diseño visual representativo. Helper `drawFloating(cv)` dibuja dos columnas de fondo + rectángulo desplazado
- `config.toml`: `window_float_toggle = "Control-Shift-g"`

## [2.7.1] - 2026-02-25

### Changed

- **Autotile: cálculo dinámico de columnas universal**: El layout autotile ahora calcula dinámicamente el número de columnas según la cantidad de ventanas abiertas en **todos** los monitores, no solo en ultrawide
- **Autotile: límite de columnas para resolución estándar**: En monitores no-ultrawide (< `ultrawide_threshold`), el máximo de columnas se limita a 2 para evitar ventanas demasiado estrechas
- **Atajo toggle**: Cambiado de `Control-Shift-T` a `Control-Shift-Y` para evitar conflictos con atajos comunes
- **Atajo restore**: Cambiado de `Control-Shift-R` a `Control-Shift-U` para evitar conflictos con atajos comunes
- **Versión externalizada**: La versión del binario se lee desde el archivo `VERSION` en la raíz del proyecto (via `//go:embed`), eliminando el hardcoding en `main.go`

### Technical Details

- `layout/autotile.go` → `Apply()`: `calculateColumns(csize)` se invoca siempre; el cap de 2 columnas solo aplica si `!isUltrawide`
- `config.toml` → `[keys]`: `toggle = "Control-Shift-Y"`
- `VERSION`: nuevo archivo en la raíz con la versión semántica del binario
- `main.go`: versión leída con `//go:embed VERSION` en lugar de string literal

## [2.7.0] - 2026-02-25

### Added

- **Proporciones por columna en autotile**: Nueva funcionalidad para ajustar proporciones individuales de columnas
- **Atajos de proporción**: Ctrl+Shift+KP_3 (increase) y Ctrl+Shift+KP_1 (decrease) ahora funcionan en autotile
- **Reset automático**: Las proporciones se resetean al activar/desactivar autotile o tiling
- **Distribución equilibrada**: Columnas centrales se expanden hacia ambos lados, no solo hacia uno

### Changed

- **Algoritmo de distribución**: Columnas ajustan espacio de vecinos de manera inteligente
- **Reset de layout**: Mejorado el reset de proporciones al cambiar layouts

### Fixed

- **Interfaz desordenada**: Corregido problema donde columnas centrales solo se expandían hacia un lado
- **Consistencia**: Asegurado que todas las columnas mantengan proporciones válidas

### Technical Details

- **Nuevos métodos**: `ResetColumnProportions()`, `adjustActiveColumnProportion()`, `normalizeColumnProportions()`
- **Campo nuevo**: `ColumnProps []float64` en `AutotileLayout` para proporciones por columna
- **Algoritmo mejorado**: Distribución 50%/50% para columnas centrales entre vecinos izquierdo/derecho
- **Archivos actualizados**:
  - `layout/autotile.go`: Implementación completa de proporciones por columna
  - `input/action.go`: Reset automático en `EnableTiling()` y `AutotileLayoutAction()`

## [2.6.0] - 2026-02-25

### Added

- **Autotile layout**: New dynamic column-based layout for ultrawide displays
- **Ultrawide detection**: Automatic layout switching based on screen width threshold
- **Column management**: Dynamic column adjustment with keyboard shortcuts
- **New keyboard shortcuts**:
  - `Control-Shift-A`: Toggle between autotile and vertical-left layouts
  - `Control-Shift-KP_Multiply`: Increase number of columns in autotile
  - `Control-Shift-KP_Divide`: Decrease number of columns in autotile
- **Configuration options**:
  - `ultrawide_threshold`: Screen width to trigger autotile mode (default: 2560px)
  - `autotile_columns_max`: Maximum columns for autotile layout (default: 4)
  - `autotile_columns_default`: Default columns for autotile layout (default: 4)
- **Installation script**: Complete update script with systemd service management

### Changed

- **Default layout**: Changed from `vertical-right` to `autotile`
- **Layout cycle**: Added `autotile` to default layout cycle
- **Interface extension**: All layouts now implement column management methods
- **Configuration**: Updated `config.toml` with new autotile settings
- **Version**: Updated to 2.6.0 for fork with autotile feature

### Fixed

- **Code structure**: Consistent interface implementation across all layouts
- **Icon rendering**: Added autotile icon representation in systray

### Technical Details

- **New file**: `layout/autotile.go` - Complete autotile implementation
- **Updated files**:
  - `common/config.go`: Added configuration struct fields
  - `config.toml`: Updated with autotile configuration
  - `desktop/layout.go`: Extended Layout interface
  - `desktop/workspace.go`: Added autotile to layout creation
  - `input/action.go`: New autotile actions and shortcuts
  - `layout/fullscreen.go`, `layout/horizontal.go`, `layout/maximized.go`, `layout/vertical.go`: Interface compliance
  - `ui/icon.go`: Autotile icon rendering
- **Script**: `update-cortile.sh` - Complete installation and update script

## [2.5.2] - 2024-02-24

### Changed

- Updated Go version

## [2.5.1] - 2024-02-24

### Changed

- Updated dependencies

## [2.5.0] - 2024-02-24

### Added

- Runtime feature flags for custom builds (#76)
- Configurable layout cycle order (#78)

## [2.4.0] - 2024-02-24

### Added

- Enhanced window management features

## [2.3.3] - 2024-02-24

### Fixed

- Bug fixes and stability improvements

## [2.3.2] - 2024-02-24

### Fixed

- Performance optimizations

## [2.3.1] - 2024-02-24

### Fixed

- Minor bug fixes

## [2.3.0] - 2024-02-24

### Added

- New window management features

## [2.2.2] - 2024-02-24

### Fixed

- Stability improvements

## [2.2.1] - 2024-02-24

### Fixed

- Bug fixes

## [2.2.0] - 2024-02-24

### Added

- Enhanced layout management

## [2.1.1] - 2024-02-24

### Fixed

- Minor fixes

## [2.1.0] - 2024-02-24

### Added

- Initial feature set for v2.1

## [2.0.0] - 2024-02-24

### Added

- Major rewrite and new architecture
- Complete refactoring of codebase

## [1.2.0] - 2024-02-24

### Added

- Advanced features for v1.x series

## [1.1.5] - 2024-02-24

### Fixed

- Stability improvements

## [1.1.4] - 2024-02-24

### Fixed

- Bug fixes

## [1.1.3] - 2024-02-24

### Fixed

- Performance optimizations

## [1.1.2] - 2024-02-24

### Fixed

- Minor fixes

## [1.1.1] - 2024-02-24

### Fixed

- Initial bug fixes for v1.1

## [1.1.0] - 2024-02-24

### Added

- New features for v1.1 release

## [1.0.0-rc.3] - 2024-02-24

### Changed

- Release candidate 3

## [1.0.0-rc.2] - 2024-02-24

### Changed

- Release candidate 2

## [1.0.0-rc.1] - 2024-02-24

### Changed

- Release candidate 1

## [1.0.0b3] - 2024-02-24

### Changed

- Beta 3 release

## [1.0.0b2] - 2024-02-24

### Changed

- Beta 2 release

## [1.0.0b1] - 2024-02-24

### Changed

- Beta 1 release

## [1.0.0a4] - 2024-02-24

### Changed

- Alpha 4 release

## [1.0.0a3] - 2024-02-24

### Changed

- Alpha 3 release

## [1.0.0a2] - 2024-02-24

### Changed

- Alpha 2 release

## [1.0.0a1] - 2024-02-24

### Changed

- Alpha 1 release

## [1.0.0] - 2024-02-24

### Added

- Initial stable release of Cortile with core tiling functionality
- Workspace based tiling
- Auto detection of panels
- Toggle window decorations
- User interface for tiling mode
- Systray icon indicator and menu
- Custom addons via python bindings
- Keyboard, hot corner and systray bindings
- Vertical, horizontal, maximized and fullscreen mode
- Remember layout proportions
- Floating and sticky windows
- Drag & drop window swap
- Workplace aware layouts
- Multi monitor support
