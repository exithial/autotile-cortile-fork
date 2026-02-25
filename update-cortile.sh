#!/bin/bash

# Script para compilar, instalar y configurar cortile con soporte autotile
# Reemplaza completamente config.toml y maneja servicios systemd
# Uso: ./update-cortile-final.sh [--install-system-wide]

set -e

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directorios
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build"
BIN_NAME="cortile"
USER_BIN_DIR="$HOME/.local/bin"
SYSTEM_BIN_DIR="/usr/local/bin"
CONFIG_DIR="$HOME/.config/cortile"
CONFIG_FILE="$CONFIG_DIR/config.toml"
BACKUP_DIR="$CONFIG_DIR/backup"
SERVICE_DIR="$HOME/.config/systemd/user"
SERVICE_FILE="$SERVICE_DIR/cortile.service"

# Variables de control
INSTALL_SYSTEM_WIDE=false
FORCE=false
BACKUP=true
RESTART_SERVICE=true

# Parsear argumentos
while [[ $# -gt 0 ]]; do
    case $1 in
        --install-system-wide)
            INSTALL_SYSTEM_WIDE=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        --no-backup)
            BACKUP=false
            shift
            ;;
        --no-restart)
            RESTART_SERVICE=false
            shift
            ;;
        -h|--help)
            echo "Uso: $0 [OPCIONES]"
            echo ""
            echo "Opciones:"
            echo "  --install-system-wide  Instalar en /usr/local/bin (requiere sudo)"
            echo "  --force                Forzar recompilación sin preguntas"
            echo "  --no-backup            No crear backup de configuración"
            echo "  --no-restart           No reiniciar servicio systemd"
            echo "  -h, --help             Mostrar esta ayuda"
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Argumento desconocido '$1'${NC}"
            exit 1
            ;;
    esac
done

# Función para imprimir mensajes
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar dependencias
check_dependencies() {
    log_info "Verificando dependencias..."
    
    if ! command -v go &> /dev/null; then
        log_error "Go no está instalado. Instala Go 1.22 o superior."
        exit 1
    fi
    
    GO_VERSION=$(go version | grep -oP 'go\d+\.\d+')
    if [[ "$GO_VERSION" < "go1.22" ]]; then
        log_error "Se requiere Go 1.22 o superior. Versión actual: $GO_VERSION"
        exit 1
    fi
    
    log_success "Go $GO_VERSION detectado"
}

# Detener procesos cortile en ejecución
stop_cortile_processes() {
    log_info "Deteniendo procesos cortile..."
    
    # Buscar solo procesos cortile reales (no rutas que contengan 'cortile')
    # Buscar por nombre exacto del binario o procesos que ejecutan /usr/local/bin/cortile o ~/.local/bin/cortile
    PIDS=""
    
    # Buscar procesos que ejecutan el binario cortile
    for BIN_PATH in "/usr/local/bin/cortile" "$USER_BIN_DIR/cortile" "$SYSTEM_BIN_DIR/cortile"; do
        if [[ -f "$BIN_PATH" ]]; then
            # Buscar procesos ejecutando este binario específico
            BIN_PIDS=$(pgrep -f "^$BIN_PATH" 2>/dev/null || true)
            if [[ -n "$BIN_PIDS" ]]; then
                PIDS="$PIDS $BIN_PIDS"
            fi
        fi
    done
    
    # También buscar por nombre cortile sin ruta (para procesos ya ejecutándose)
    CORTILE_PIDS=$(pgrep -x "cortile" 2>/dev/null || true)
    if [[ -n "$CORTILE_PIDS" ]]; then
        PIDS="$PIDS $CORTILE_PIDS"
    fi
    
    # Eliminar duplicados y espacios extra
    PIDS=$(echo "$PIDS" | tr ' ' '\n' | sort -u | tr '\n' ' ' | sed 's/^[[:space:]]*//;s/[[:space:]]*$//')
    
    if [[ -n "$PIDS" ]]; then
        log_info "Procesos cortile encontrados: $PIDS"
        kill $PIDS 2>/dev/null || true
        sleep 1
        # Verificar si aún están ejecutándose
        STILL_RUNNING=""
        for PID in $PIDS; do
            if kill -0 "$PID" 2>/dev/null; then
                STILL_RUNNING="$STILL_RUNNING $PID"
            fi
        done
        
        if [[ -n "$STILL_RUNNING" ]]; then
            log_warning "Algunos procesos aún ejecutándose, forzando terminación..."
            kill -9 $STILL_RUNNING 2>/dev/null || true
        fi
        
        log_success "Procesos cortile detenidos"
    else
        log_info "No hay procesos cortile en ejecución"
    fi
}

# Detener servicio systemd si existe
stop_systemd_service() {
    log_info "Verificando servicio systemd..."
    
    # Verificar si systemd está disponible
    if ! command -v systemctl &> /dev/null; then
        log_info "systemctl no disponible, saltando gestión de servicios"
        return
    fi
    
    # Verificar y detener servicios cortile
    if systemctl --user is-active cortile.service &>/dev/null 2>&1; then
        log_info "Deteniendo servicio cortile..."
        systemctl --user stop cortile.service
        log_success "Servicio cortile detenido"
    fi
    
    if systemctl --user is-active cortile-autotile.service &>/dev/null 2>&1; then
        log_info "Deteniendo servicio cortile-autotile..."
        systemctl --user stop cortile-autotile.service
        log_success "Servicio cortile-autotile detenido"
    fi
}

# Crear directorios
create_directories() {
    log_info "Creando directorios..."
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$USER_BIN_DIR"
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$SERVICE_DIR"
    
    if $BACKUP; then
        mkdir -p "$BACKUP_DIR"
    fi
    
    log_success "Directorios creados"
}

# Backup de configuración existente
backup_config() {
    if $BACKUP && [[ -f "$CONFIG_FILE" ]]; then
        log_info "Creando backup de configuración..."
        
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        BACKUP_FILE="$BACKUP_DIR/config.toml.$TIMESTAMP.backup"
        
        cp "$CONFIG_FILE" "$BACKUP_FILE"
        log_success "Backup creado: $BACKUP_FILE"
    fi
}

# Compilar cortile
compile_cortile() {
    log_info "Compilando cortile..."
    
    cd "$SCRIPT_DIR"
    
    # Verificar si hay cambios sin commit
    if git status --porcelain | grep -q "^ M"; then
        log_warning "Hay cambios sin commit en el repositorio"
        if ! $FORCE; then
            read -p "¿Continuar de todos modos? (s/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Ss]$ ]]; then
                log_info "Compilación cancelada"
                exit 0
            fi
        fi
    fi
    
    # Limpiar build anterior
    rm -rf "$BUILD_DIR"
    mkdir -p "$BUILD_DIR"
    
    # Compilar
    GOOS=$(go env GOOS)
    GOARCH=$(go env GOARCH)
    
    log_info "Compilando para $GOOS/$GOARCH..."
    
    go build -ldflags="-X 'main.date=$(date --iso-8601=seconds)'" \
             -o "$BUILD_DIR/$BIN_NAME" .
    
    if [[ $? -ne 0 ]]; then
        log_error "Error durante la compilación"
        exit 1
    fi
    
    # Verificar binario
    if [[ ! -f "$BUILD_DIR/$BIN_NAME" ]]; then
        log_error "Binario no generado"
        exit 1
    fi
    
    # Dar permisos de ejecución
    chmod +x "$BUILD_DIR/$BIN_NAME"
    
    log_success "Compilación completada: $BUILD_DIR/$BIN_NAME"
}

# Instalar binario
install_binary() {
    log_info "Instalando binario..."
    
    if $INSTALL_SYSTEM_WIDE; then
        # Instalar system-wide
        if [[ $EUID -ne 0 ]]; then
            log_error "Se requieren privilegios de superusuario para instalar en $SYSTEM_BIN_DIR"
            log_info "Ejecuta: sudo $0 --install-system-wide"
            exit 1
        fi
        
        # Backup de binario existente
        if [[ -f "$SYSTEM_BIN_DIR/$BIN_NAME" ]]; then
            TIMESTAMP=$(date +%Y%m%d_%H%M%S)
            cp "$SYSTEM_BIN_DIR/$BIN_NAME" "$SYSTEM_BIN_DIR/$BIN_NAME.$TIMESTAMP.backup"
            log_success "Backup de binario creado"
        fi
        
        # Copiar nuevo binario
        cp "$BUILD_DIR/$BIN_NAME" "$SYSTEM_BIN_DIR/$BIN_NAME"
        chmod +x "$SYSTEM_BIN_DIR/$BIN_NAME"
        
        BIN_PATH="$SYSTEM_BIN_DIR/$BIN_NAME"
        log_success "Instalado en $BIN_PATH"
    else
        # Instalar user-local
        cp "$BUILD_DIR/$BIN_NAME" "$USER_BIN_DIR/$BIN_NAME"
        chmod +x "$USER_BIN_DIR/$BIN_NAME"
        
        BIN_PATH="$USER_BIN_DIR/$BIN_NAME"
        
        # Verificar si $USER_BIN_DIR está en PATH
        if [[ ":$PATH:" != *":$USER_BIN_DIR:"* ]]; then
            log_warning "$USER_BIN_DIR no está en PATH"
            log_info "Agrega esto a tu ~/.bashrc o ~/.zshrc:"
            echo "export PATH=\"\$PATH:$USER_BIN_DIR\""
        fi
        
        log_success "Instalado en $BIN_PATH"
    fi
    
    # Guardar ruta del binario para uso posterior
    echo "$BIN_PATH" > /tmp/cortile_bin_path
}

# Reemplazar completamente config.toml
replace_config() {
    log_info "Reemplazando configuración completa..."
    
    # Usar el config.toml del repositorio (ya actualizado con autotile)
    cp "$SCRIPT_DIR/config.toml" "$CONFIG_FILE"
    
    log_success "Configuración reemplazada: $CONFIG_FILE"
}

# Configurar servicio systemd
setup_systemd_service() {
    log_info "Configurando servicio systemd..."
    
    # Leer ruta del binario
    BIN_PATH=$(cat /tmp/cortile_bin_path 2>/dev/null || echo "")
    if [[ -z "$BIN_PATH" ]]; then
        if $INSTALL_SYSTEM_WIDE; then
            BIN_PATH="$SYSTEM_BIN_DIR/$BIN_NAME"
        else
            BIN_PATH="$USER_BIN_DIR/$BIN_NAME"
        fi
    fi
    
    # Verificar si el servicio ya existe
    if [[ -f "$SERVICE_FILE" ]]; then
        log_info "Servicio existente encontrado, actualizando..."
        # Backup del servicio existente
        TIMESTAMP=$(date +%Y%m%d_%H%M%S)
        cp "$SERVICE_FILE" "$SERVICE_FILE.$TIMESTAMP.backup"
        log_success "Backup de servicio creado"
    else
        log_info "Creando nuevo servicio..."
    fi
    
    # Crear archivo de servicio
    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=Cortile Auto Tiling Manager
After=graphical-session.target
PartOf=graphical-session.target

[Service]
Type=simple
ExecStart=$BIN_PATH
Restart=on-failure
RestartSec=5
Environment="DISPLAY=:0"
Environment="XAUTHORITY=%h/.Xauthority"
Environment="XDG_RUNTIME_DIR=/run/user/%U"

[Install]
WantedBy=default.target
EOF
    
    # Recargar systemd si está disponible
    if command -v systemctl &> /dev/null; then
        systemctl --user daemon-reload
        systemctl --user enable cortile.service
        log_success "Servicio configurado y habilitado: $SERVICE_FILE"
    else
        log_warning "systemctl no disponible, servicio creado pero no habilitado"
        log_success "Archivo de servicio creado: $SERVICE_FILE"
    fi
}

# Iniciar cortile (con o sin systemd)
start_cortile() {
    if $RESTART_SERVICE; then
        log_info "Iniciando cortile..."
        
        # Leer ruta del binario
        BIN_PATH=$(cat /tmp/cortile_bin_path 2>/dev/null || echo "")
        if [[ -z "$BIN_PATH" ]]; then
            if $INSTALL_SYSTEM_WIDE; then
                BIN_PATH="$SYSTEM_BIN_DIR/$BIN_NAME"
            else
                BIN_PATH="$USER_BIN_DIR/$BIN_NAME"
            fi
        fi
        
        # Verificar si systemd está disponible
        if command -v systemctl &> /dev/null; then
            log_info "Usando systemd para iniciar..."
            systemctl --user start cortile.service
            sleep 2
            
            if systemctl --user is-active cortile.service &>/dev/null; then
                log_success "Servicio iniciado correctamente via systemd"
            else
                log_warning "No se pudo iniciar via systemd, intentando manualmente..."
                nohup "$BIN_PATH" > /dev/null 2>&1 &
                log_success "Cortile iniciado manualmente en background"
            fi
        else
            log_info "systemctl no disponible, iniciando manualmente..."
            nohup "$BIN_PATH" > /dev/null 2>&1 &
            log_success "Cortile iniciado manualmente en background"
        fi
    else
        log_info "Inicio automático deshabilitado (--no-restart)"
    fi
}

# Verificar instalación
verify_installation() {
    log_info "Verificando instalación..."
    
    # Leer ruta del binario
    BIN_PATH=$(cat /tmp/cortile_bin_path 2>/dev/null || echo "")
    if [[ -z "$BIN_PATH" ]]; then
        if $INSTALL_SYSTEM_WIDE; then
            BIN_PATH="$SYSTEM_BIN_DIR/$BIN_NAME"
        else
            BIN_PATH="$USER_BIN_DIR/$BIN_NAME"
        fi
    fi
    
    if [[ ! -f "$BIN_PATH" ]]; then
        log_error "Binario no encontrado en $BIN_PATH"
        exit 1
    fi
    
    # Verificar versión
    log_info "Probando binario..."
    "$BIN_PATH" --version 2>&1 | head -1
    
    # Verificar configuración
    if [[ ! -f "$CONFIG_FILE" ]]; then
        log_error "Configuración no encontrada en $CONFIG_FILE"
        exit 1
    fi
    
    # Verificar que tenga configuraciones autotile
    if ! grep -q "autotile" "$CONFIG_FILE"; then
        log_error "Configuración autotile no encontrada"
        exit 1
    fi
    
    log_success "Instalación verificada"
}

# Mostrar resumen
show_summary() {
    echo ""
    echo -e "${GREEN}========================================${NC}"
    echo -e "${GREEN}         INSTALACIÓN COMPLETADA         ${NC}"
    echo -e "${GREEN}========================================${NC}"
    echo ""
    
    # Leer ruta del binario
    BIN_PATH=$(cat /tmp/cortile_bin_path 2>/dev/null || echo "")
    if [[ -z "$BIN_PATH" ]]; then
        if $INSTALL_SYSTEM_WIDE; then
            BIN_PATH="$SYSTEM_BIN_DIR/$BIN_NAME"
        else
            BIN_PATH="$USER_BIN_DIR/$BIN_NAME"
        fi
    fi
    
    echo -e "${BLUE}Binario instalado en:${NC} $BIN_PATH"
    
    if [[ "$BIN_PATH" == "$USER_BIN_DIR/$BIN_NAME" ]] && [[ ":$PATH:" != *":$USER_BIN_DIR:"* ]]; then
        echo -e "${YELLOW}Nota:${NC} Agrega esto a tu shell:"
        echo "  export PATH=\"\$PATH:$USER_BIN_DIR\""
    fi
    
    echo -e "${BLUE}Ejecutar como:${NC} cortile"
    echo ""
    echo -e "${BLUE}Configuración en:${NC} $CONFIG_FILE"
    echo -e "${BLUE}Backups en:${NC} $BACKUP_DIR"
    echo -e "${BLUE}Servicio systemd:${NC} $SERVICE_FILE"
    echo ""
    echo -e "${GREEN}Nuevas funcionalidades:${NC}"
    echo "  • Layout 'autotile' para pantallas ultrawide"
    echo "  • Único atajo: Control-Shift-A (toggle autotile/vertical-left)"
    echo "  • Atajos: Control-Shift-KP_Multiply/Divide (ajustar columnas)"
    echo "  • Configuración completa reemplazada"
    echo ""
    
    if $RESTART_SERVICE; then
        echo -e "${YELLOW}Estado:${NC}"
        if command -v systemctl &> /dev/null && systemctl --user is-active cortile.service &>/dev/null 2>&1; then
            echo "  Servicio systemd activo"
            systemctl --user status cortile.service --no-pager | head -5
        else
            # Verificar si está ejecutándose
            PIDS=$(pgrep -f "cortile" 2>/dev/null || true)
            if [[ -n "$PIDS" ]]; then
                echo "  Cortile ejecutándose en background"
                echo "  PID: $PIDS"
            else
                echo "  Cortile no está ejecutándose"
                echo "  Para iniciar: cortile"
            fi
        fi
    else
        echo -e "${YELLOW}Para iniciar manualmente:${NC}"
        echo "  cortile"
        if command -v systemctl &> /dev/null; then
            echo "  # o con systemd:"
            echo "  systemctl --user start cortile.service"
        fi
    fi
    echo ""
}

# Limpiar archivos temporales
cleanup() {
    rm -f /tmp/cortile_bin_path 2>/dev/null || true
}

# Función principal
main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE}    ACTUALIZACIÓN COMPLETA DE CORTILE   ${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    trap cleanup EXIT
    
    check_dependencies
    stop_cortile_processes
    stop_systemd_service
    create_directories
    backup_config
    compile_cortile
    install_binary
    replace_config
    setup_systemd_service
    start_cortile
    verify_installation
    show_summary
}

# Ejecutar
main "$@"