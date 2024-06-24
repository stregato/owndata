package core

import "path"

func Map[K comparable, T any](s []T, f func(T) K) map[K]T {
	result := make(map[K]T)

	for _, v := range s {
		result[f(v)] = v
	}
	return result
}

func Keys[T comparable, U any](s map[T]U) []T {
	var result []T
	for v := range s {
		result = append(result, v)
	}
	return result
}

func Values[T comparable, U any](s map[T]U) []U {
	var result []U
	for _, v := range s {
		result = append(result, v)
	}
	return result
}

func Contains[T comparable](s []T, v T) bool {
	for _, e := range s {
		if e == v {
			return true
		}
	}
	return false
}

func SplitPath(p string) (string, string) {
	dir, base := path.Split(p)
	dir = path.Clean(dir)
	if dir == "." {
		dir = ""
	}
	return dir, base
}

func Dir(p string) string {
	dir := path.Dir(p)
	dir = path.Clean(dir)
	if dir == "." {
		dir = ""
	}
	return dir
}
