# The control file has a few required fields
# name, version, architecture, maintainer, and description

VERSION=$CIRCLE_TAG                           # version of the application
REVISION=1                                    # revision of the current deb package
ARCH=amd64                                    # hardware architecture
MAINTAINER="Cory Schwartz <cory@protocol.ai>" # package maintainer


LOTUS="lotus_${VERSION}_${REVISION}_${ARCH}"
LOTUS_DAEMON="lotus-daemon_${VERSION}_${REVISION}_${ARCH}"

# make the deb control file.
function mkcontrol() {
	WORKDIR=$1
	PACKAGE=$2
	DESCRIPTION=$3
	DEPENDS=$4

	mkdir "${WORKDIR}/DEBIAN"
	cat >"${WORKDIR}/DEBIAN/control" <<EOF
Package: "${PACKAGE}"
Version: "${VERSION}"
Architecture: "${ARCH}" 
Maintainer: "${MAINTAINER}"
Description: ${DESCRIPTION}"
Depends: "${DEPENDS}"
EOF
}

# lotus binary package
mkdir -p "${LOTUS}/usr/bin"
cp linux/lotus "${LOTUS}/usr/bin/"
chmod +x "${LOTUS}/usr/bin/lotus"
mkcontrol "${LOTUS}" lotus "hwloc, ocl-icd-libopencl1"
dpkg-deb --build --root-owner-group  "${LOTUS}"
