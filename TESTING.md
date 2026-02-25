# Testing

Este proyecto no cuenta con una suite de tests automatizados (`*_test.go`).
La verificación se realiza mediante pruebas manuales funcionales.

---

## Historial de Verificación

### [2026-02-25] — v2.7.1

**Estado:** ✅ VERIFICADO MANUALMENTE

**Cambios verificados:**

| Escenario                | Resultado                            |
| ------------------------ | ------------------------------------ |
| 1 ventana en 1920×1080   | Pantalla completa ✅                 |
| 2 ventanas en 1920×1080  | 2 columnas iguales ✅                |
| 3 ventanas en 1920×1080  | 2 columnas (col izq. con 2 filas) ✅ |
| 1 ventana en ultrawide   | Pantalla completa ✅                 |
| 4 ventanas en ultrawide  | 4 columnas ✅                        |
| Atajo `Ctrl+Shift+Y`     | Toggle enable/disable ✅             |
| Compilación (`go build`) | Sin errores ✅                       |
| Servicio systemd         | Activo y estable ✅                  |

**Método:** Compilación y despliegue con `./update-cortile.sh`, verificación visual en sesión X11.

---

### [2026-02-25] — v2.7.0

**Estado:** ✅ VERIFICADO MANUALMENTE

**Cambios verificados:**

| Escenario                           | Resultado                              |
| ----------------------------------- | -------------------------------------- |
| Proporciones por columna (autotile) | Ajuste y persistencia correctos ✅     |
| Atajos `Ctrl+Shift+KP_3` / `KP_1`   | Incremento/decremento de proporción ✅ |
| Reset automático al cambiar layout  | Proporciones reseteadas ✅             |
| Compilación (`go build`)            | Sin errores ✅                         |

---

### [2026-02-25] — v2.6.0

**Estado:** ✅ VERIFICADO MANUALMENTE

**Cambios verificados:**

| Escenario                              | Resultado               |
| -------------------------------------- | ----------------------- |
| Layout autotile en ultrawide           | Columnas dinámicas ✅   |
| Atajo `Ctrl+Shift+A` (toggle autotile) | Funcional ✅            |
| Atajo `Ctrl+Shift+KP_Multiply`         | Aumenta columnas ✅     |
| Atajo `Ctrl+Shift+KP_Divide`           | Disminuye columnas ✅   |
| Script `update-cortile.sh`             | Instalación completa ✅ |
