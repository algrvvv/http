package logger

func Red(s string) string {
	return "\033[0;31m" + s + "\033[0m"
}

func Green(s string) string {
	return "\033[0;32m" + s + "\033[0m"
}

func Yellow(s string) string {
	return "\033[0;33m" + s + "\033[0m"
}

func Blue(s string) string {
	return "\033[0;34m" + s + "\033[0m"
}

func LightBlue(s string) string {
	return "\033[94m" + s + "\033[0m"
}

func LightGreen(s string) string {
	return "\033[92m" + s + "\033[0m"
}

func Orange(s string) string {
	return "\033[93m" + s + "\033[0m"
}

func LightRed(s string) string {
	return "\033[91m" + s + "\033[0m"
}
