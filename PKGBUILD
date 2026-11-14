# Maintainer: nagnagi
pkgname=wrep
pkgver=0.1.1
pkgrel=1
pkgdesc='List and switch Hyprland windows via dmenu-compatible launchers (fuzzel, rofi, wofi, dmenu)'
arch=('x86_64' 'aarch64')
url='https://github.com/nagnagi/wrep'
license=('MIT')
depends=('hyprland')
makedepends=('go' 'git')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/nagnagi/wrep/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=('d98b0ac63b7c61c2035bbe6bd8553fc375b2bef2817568959349156046790443')

build() {
    cd "${srcdir}/${pkgname}-${pkgver}"
    export CGO_ENABLED=0
    export GOFLAGS='-trimpath -buildvcs=false'
    go build -o "${pkgname}" .
}

check() {
    cd "${srcdir}/${pkgname}-${pkgver}"
    go test ./...
}

package() {
    cd "${srcdir}/${pkgname}-${pkgver}"
    install -Dm755 "${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
    install -Dm644 LICENSE "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
    install -Dm644 README.md "${pkgdir}/usr/share/doc/${pkgname}/README.md"
}
