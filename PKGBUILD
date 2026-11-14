# Maintainer: nagnagi
pkgname=wrep
pkgver=0.1.0
pkgrel=1
pkgdesc='List and switch Hyprland windows via dmenu-compatible launchers (fuzzel, rofi, wofi, dmenu)'
arch=('x86_64' 'aarch64')
url='https://github.com/nagnagi/wrep'
license=('MIT')
depends=('hyprland')
makedepends=('go' 'git')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/nagnagi/wrep/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=('SKIP')

pkgver() {
    cd "${srcdir}/${pkgname}"
    if git describe --long --tags --match 'v*' &>/dev/null; then
        git describe --long --tags --match 'v*' \
            | sed 's/^v//; s/\([^-]*-g\)/r\1/; s/-/./g'
    else
        echo "0.1.0.r$(git rev-list --count HEAD).g$(git rev-parse --short HEAD)"
    fi
}

build() {
    cd "${srcdir}/${pkgname}"
    export CGO_ENABLED=0
    export GOFLAGS='-trimpath -buildvcs=false'
    go build -o "${pkgname}" .
}

check() {
    cd "${srcdir}/${pkgname}"
    go test ./...
}

package() {
    cd "${srcdir}/${pkgname}"
    install -Dm755 "${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
    install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
    install -Dm644 README.md "${pkgdir}/usr/share/doc/${pkgname}/README.md"
}
