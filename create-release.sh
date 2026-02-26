#!/usr/bin/env bash
# =============================================================================
#  create-release.sh — Crea un release de GitHub con el binario compilado
#
#  Uso: ./create-release.sh
#
#  Requisitos:
#    - gh (GitHub CLI) autenticado: gh auth login
#    - go >= 1.22
#
#  Lee VERSION y release_notes.txt automáticamente.
# =============================================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# ─── Colores ──────────────────────────────────────────────────────────────────
RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'
NC='\033[0m'

info()    { echo -e "${BLUE}[INFO]${NC} $*"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $*"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $*"; }
error()   { echo -e "${RED}[ERROR]${NC} $*"; exit 1; }

# ─── Verificar dependencias ───────────────────────────────────────────────────
info "Verificando dependencias..."
command -v go  >/dev/null 2>&1 || error "go no encontrado. Instala Go >= 1.22"
command -v gh  >/dev/null 2>&1 || error "gh CLI no encontrado. Instala con: sudo apt install gh"
gh auth status >/dev/null 2>&1 || error "gh no está autenticado. Ejecuta: gh auth login"

# ─── Leer versión y notas ─────────────────────────────────────────────────────
VERSION_FILE="$SCRIPT_DIR/VERSION"
NOTES_FILE="$SCRIPT_DIR/release_notes.txt"

[[ -f "$VERSION_FILE" ]] || error "No se encontró VERSION"
[[ -f "$NOTES_FILE"   ]] || error "No se encontró release_notes.txt"

VERSION="v$(cat "$VERSION_FILE" | tr -d '[:space:]')"
info "Versión detectada: $VERSION"

# ─── Detectar owner/repo desde el remote ─────────────────────────────────────
REMOTE_URL=$(git remote get-url origin 2>/dev/null || echo "")
if [[ "$REMOTE_URL" =~ github\.com[:/]([^/]+)/([^/.]+) ]]; then
    GH_OWNER="${BASH_REMATCH[1]}"
    GH_REPO="${BASH_REMATCH[2]}"
else
    error "No se pudo detectar el repositorio de GitHub desde el remote origin"
fi
info "Repositorio: $GH_OWNER/$GH_REPO"

# ─── Compilar binario ─────────────────────────────────────────────────────────
BUILD_DIR="$SCRIPT_DIR/build/release"
BINARY_NAME="cortile"
BINARY_PATH="$BUILD_DIR/${BINARY_NAME}_${VERSION}_linux_amd64.tar.gz"

mkdir -p "$BUILD_DIR"

info "Compilando para linux/amd64..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.version=$(cat "$VERSION_FILE" | tr -d '[:space:]')" \
    -o "$BUILD_DIR/$BINARY_NAME" \
    .
success "Binario compilado: $BUILD_DIR/$BINARY_NAME"

# Empaquetar en tar.gz junto con config y README
info "Empaquetando assets..."
tar -czf "$BINARY_PATH" \
    -C "$BUILD_DIR" "$BINARY_NAME" \
    -C "$SCRIPT_DIR" config.toml README.md
success "Archivo empaquetado: $BINARY_PATH"

# Generar checksum
CHECKSUM_FILE="$BUILD_DIR/cortile_${VERSION}_checksums.txt"
sha256sum "$BINARY_PATH" > "$CHECKSUM_FILE"
success "Checksum generado: $CHECKSUM_FILE"

# ─── Verificar si el tag existe, si no, crearlo ───────────────────────────────
if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
    info "Creando tag $VERSION..."
    git tag "$VERSION"
    git push origin "$VERSION"
    success "Tag $VERSION creado y pusheado"
else
    info "Tag $VERSION ya existe"
fi

# ─── Crear o actualizar release en GitHub ─────────────────────────────────────
RELEASE_TITLE="$(head -1 "$NOTES_FILE" | sed 's/^v[0-9.]* - //' | xargs)"

if gh release view "$VERSION" --repo "$GH_OWNER/$GH_REPO" >/dev/null 2>&1; then
    warning "Release $VERSION ya existe. Actualizando assets..."
    gh release upload "$VERSION" \
        "$BINARY_PATH" \
        "$CHECKSUM_FILE" \
        --repo "$GH_OWNER/$GH_REPO" \
        --clobber
    success "Assets actualizados en el release existente $VERSION"
else
    info "Creando release $VERSION..."
    gh release create "$VERSION" \
        "$BINARY_PATH" \
        "$CHECKSUM_FILE" \
        --repo "$GH_OWNER/$GH_REPO" \
        --title "$VERSION - $RELEASE_TITLE" \
        --notes-file "$NOTES_FILE" \
        --latest
    success "Release $VERSION creado en GitHub"
fi

# ─── Limpieza ─────────────────────────────────────────────────────────────────
rm -f "$BUILD_DIR/$BINARY_NAME"

echo ""
echo "========================================"
echo "  RELEASE $VERSION COMPLETADO"
echo "========================================"
echo ""
echo "URL: https://github.com/$GH_OWNER/$GH_REPO/releases/tag/$VERSION"
