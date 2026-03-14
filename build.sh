#!/usr/bin/env bash
set -euo pipefail

# =============================
# Параметры
# =============================
WORKSPACE="app"
BUILD_ID=${BUILD_ID:-$(date +%s)}
BUILD_NUMBER=${BUILD_NUMBER:-1}

PKG_NAME="elza-breeder"

# читаем версию из debian/changelog
PKG_VER=$(head -n1 debian/changelog | sed -E 's/^[^ ]+ \(([^)]+)\).*/\1/')

IMAGE_NAME="golang/elza-breeder"

TMP_DIR="/tmp/c-${PKG_NAME}-deb"

# SSH параметры для upload
SSH_HOST="elzabreeder.ru"
SSH_USER="user1"
REMOTE_DIR="${PKG_NAME}"
# путь на сервере, куда будет залито
REMOTE_UPLOAD_DIR="upload/${PKG_NAME}"

echo "Package: ${PKG_NAME}"
echo "Version: ${PKG_VER}"
echo "Build: ${BUILD_NUMBER}"
echo "Docker image: ${IMAGE_NAME}:b${BUILD_ID}-deb"

# =============================
# Сборка docker образа с deb
# =============================
docker build \
  -t "${IMAGE_NAME}:b${BUILD_ID}-deb" \
  -f Dockerfile.deb \
  --build-arg PKG_VER="${PKG_VER}" \
  --build-arg BUILD_NUMBER="${BUILD_NUMBER}" \
  --build-arg PKG_NAME="${PKG_NAME}" \
  .

# =============================
# Подготовка временной директории
# =============================
rm -rf "${TMP_DIR}"
mkdir -p "${TMP_DIR}"

# =============================
# Извлечение deb из контейнера
# =============================
docker rm "c-${PKG_NAME}-deb" 2>/dev/null || true
docker create --name "c-${PKG_NAME}-deb" "${IMAGE_NAME}:b${BUILD_ID}-deb"

DEB_FILE="/${PKG_NAME}_${PKG_VER}.${BUILD_NUMBER}_all.deb"

docker cp "c-${PKG_NAME}-deb:${DEB_FILE}" "${TMP_DIR}/"

docker rm "c-${PKG_NAME}-deb"

# =============================
# Создаем tar.gz для загрузки
# =============================
cd "${TMP_DIR}"
TAR_FILE="upload.tgz"
tar -czvf "${TAR_FILE}" *.deb

echo "Packaged tar: ${TAR_FILE}"

# =============================
# Upload на сервер через ssh
# =============================
# копируем архив
ssh "${SSH_USER}@${SSH_HOST}" "mkdir -p ${REMOTE_UPLOAD_DIR}"
scp "${TAR_FILE}" "${SSH_USER}@${SSH_HOST}:${REMOTE_UPLOAD_DIR}/"

# распаковываем и устанавливаем пакет
ssh "${SSH_USER}@${SSH_HOST}" bash -c "'
cd ${REMOTE_UPLOAD_DIR}
tar xf upload.tgz
rm upload.tgz
sudo dpkg -i ${PKG_NAME}_${PKG_VER}.${BUILD_NUMBER}_*.deb
# если есть неудовлетворённые зависимости, доустановить их
sudo apt-get install -f -y
rm ${PKG_NAME}_${PKG_VER}.${BUILD_NUMBER}_*.deb
'"

echo "Upload and publish finished ✅"