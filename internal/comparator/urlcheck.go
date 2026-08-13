package comparator

import (
	"fmt"
	"net"
	"net/url"
)

// ValidationError indica que la entrada del usuario no es válida (ej: URL bloqueada).
type ValidationError struct {
	message string
}

func (e *ValidationError) Error() string {
	return e.message
}

// cgnatRange cubre la red 100.64.0.0/10 (CGNAT), que net.IP.IsPrivate no considera privada.
var cgnatRange = mustParseCIDR("100.64.0.0/10")

func mustParseCIDR(s string) *net.IPNet {
	_, ipNet, err := net.ParseCIDR(s)
	if err != nil {
		panic(err)
	}
	return ipNet
}

// validateURL valida que la URL sea http/https y que el host no resuelva a direcciones
// internas (anti-SSRF).
func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return &ValidationError{message: "url inválida: " + err.Error()}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return &ValidationError{message: "solo se permiten URLs http/https"}
	}

	host := u.Hostname()
	if host == "" {
		return &ValidationError{message: "la URL debe incluir un host"}
	}

	ips, err := net.LookupIP(host)
	if err != nil {
		return &ValidationError{message: fmt.Sprintf("no se pudo resolver el host %q", host)}
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return &ValidationError{message: fmt.Sprintf("el host %q resuelve a una dirección bloqueada (%s)", host, ip)}
		}
	}

	return nil
}

// isBlockedIP indica si la IP corresponde a redes internas que no deben alcanzarse.
func isBlockedIP(ip net.IP) bool {
	return ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() ||
		ip.IsUnspecified() ||
		cgnatRange.Contains(ip)
}
