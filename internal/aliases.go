package internal

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var hostAliasPattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]{0,63}$`)

func hostAliasPath(name string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "http", "hosts", strings.ToLower(name)), nil
}

func saveHostAlias(name, target string) error {
	if !hostAliasPattern.MatchString(name) {
		return fmt.Errorf("invalid host alias: use 1–64 letters, digits, hyphens or underscores, starting with a letter or digit")
	}
	base, err := url.Parse(checkAndGetURL(target))
	if err != nil || base.Hostname() == "" || (base.Scheme != "http" && base.Scheme != "https") ||
		base.User != nil || base.RawQuery != "" || base.ForceQuery || strings.Contains(target, "#") {
		return fmt.Errorf("host alias requires an HTTP or HTTPS base URL without credentials, query or fragment")
	}
	path, err := hostAliasPath(name)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".host-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.WriteString(base.String())
	closeErr := file.Close()
	if err := errors.Join(writeErr, closeErr); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}

func resolveHostAlias(target string) (string, error) {
	end := strings.IndexAny(target, "/?#")
	if end < 0 {
		end = len(target)
	}
	name, suffix := target[:end], target[end:]
	if !hostAliasPattern.MatchString(name) {
		return checkAndGetURL(target), nil
	}
	path, err := hostAliasPath(name)
	if err != nil {
		return "", err
	}
	base, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return checkAndGetURL(target), nil
	}
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(suffix, "/") {
		return strings.TrimRight(string(base), "/") + suffix, nil
	}
	return string(base) + suffix, nil
}
