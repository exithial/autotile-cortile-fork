# Testing

Este proyecto no cuenta con una suite de tests automatizados (`*_test.go`).
La verificación se realiza mediante pruebas manuales funcionales.

---

## Historial de Verificación

### [2026-02-26] — v2.9.0

**Estado:** ⏳ PENDIENTE DE VERIFICACIÓN

**Suite de tests automatizados:** No disponible (proyecto sin `*_test.go`). Verificación funcional en sesión X11.

**Cambios verificados:**

| Escenario                                              | Resultado |
| ------------------------------------------------------ | --------- |
| Ventana flotada en monitor ultrawide (>2560px)         | Redimensiona a 1920x1080 ⏳ |
| Ventana flotada en pantalla estándar (<=2560px)         | Redimensiona al 80% ⏳ |
| Ventana flotada se centra en pantalla                  | Posición centrada ⏳ |
| Ventana plasmashell de KDE ignorada por tiling         | No aparece en layout ⏳ |
| Compilación (`go build ./...`)                         | Sin errores ⏳ |

**Método:** Compilación y despliegue con `./update-cortile.sh`, verificación visual en sesión X11.

---

### [2026-02-25] — v2.8.0

**Estado:** ✅ VERIFICADO MANUALMENTE

**Suite de tests automatizados:** No disponible (proyecto sin `*_test.go`). Verificación funcional en sesión X11.

**Cambios verificados:**

| Escenario                                         | Resultado                                       |
| ------------------------------------------------- | ----------------------------------------------- |
| `Ctrl+Shift+G` sobre ventana tileada              | Ventana sale del tiling y queda flotante ✅     |
| `Ctrl+Shift+G` sobre ventana flotante             | Ventana vuelve al tiling en layout activo ✅    |
| Overlay `floating` al activar                     | Diagrama representativo centrado en pantalla ✅ |
| Overlay `tiling` al desactivar                    | Layout actual con clientes dibujados ✅         |
| Re-tile forzado tras unfloat                      | Ventana reintegrada sin reiniciar Cortile ✅    |
| Múltiples ventanas: flotar una, resto se re-tilea | Redistributción correcta del espacio ✅         |
| Compilación (`go build ./...`)                    | Sin errores ✅                                  |
| Servicio systemd tras despliegue                  | Activo y estable ✅                             |

**Método:** Compilación y despliegue con `./update-cortile.sh`, verificación visual en sesión X11.

---

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
