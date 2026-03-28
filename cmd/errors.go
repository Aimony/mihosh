package cmd

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

const (
	exitCodeOK        = 0
	exitCodeGeneral   = 1
	exitCodeParameter = 2
	exitCodeConfig    = 3
	exitCodeNetwork   = 4
)

type commandErrorKind int

const (
	commandErrorGeneral commandErrorKind = iota
	commandErrorParameter
	commandErrorConfig
	commandErrorNetwork
)

type commandError struct {
	kind commandErrorKind
	err  error
}

func (e *commandError) Error() string {
	return e.err.Error()
}

func (e *commandError) Unwrap() error {
	return e.err
}

func wrapParameterError(err error) error {
	if err == nil {
		return nil
	}
	return &commandError{kind: commandErrorParameter, err: err}
}

func wrapConfigError(err error) error {
	if err == nil {
		return nil
	}
	return &commandError{kind: commandErrorConfig, err: err}
}

func wrapNetworkError(err error) error {
	if err == nil {
		return nil
	}
	return &commandError{kind: commandErrorNetwork, err: err}
}

func exitCodeForError(err error) int {
	switch inferErrorKind(err) {
	case commandErrorParameter:
		return exitCodeParameter
	case commandErrorConfig:
		return exitCodeConfig
	case commandErrorNetwork:
		return exitCodeNetwork
	default:
		return exitCodeGeneral
	}
}

func renderCommandError(err error) string {
	detail := strings.TrimSpace(err.Error())
	switch inferErrorKind(err) {
	case commandErrorParameter:
		return fmt.Sprintf("参数错误: %s\n使用 --help 查看命令帮助。", detail)
	case commandErrorConfig:
		return fmt.Sprintf("配置错误: %s\n可运行 `mihosh config init` 重新初始化配置。", detail)
	case commandErrorNetwork:
		return fmt.Sprintf("网络错误: %s\n请检查 API 地址、密钥和网络连通性。", detail)
	default:
		return fmt.Sprintf("执行失败: %s", detail)
	}
}

func inferErrorKind(err error) commandErrorKind {
	if err == nil {
		return commandErrorGeneral
	}

	var typed *commandError
	if errors.As(err, &typed) {
		return typed.kind
	}

	if isLikelyParameterError(err) {
		return commandErrorParameter
	}
	if isLikelyNetworkError(err) {
		return commandErrorNetwork
	}

	return commandErrorGeneral
}

func isLikelyParameterError(err error) bool {
	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	needles := []string{
		"unknown command",
		"unknown flag",
		"flag needs an argument",
		"accepts ",
		"requires at least",
		"requires at most",
		"requires exactly",
		"invalid argument",
		"参数格式错误",
		"不支持的输出格式",
	}
	for _, needle := range needles {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}

func isLikelyNetworkError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return true
	}

	msg := strings.ToLower(strings.TrimSpace(err.Error()))
	needles := []string{
		"connection refused",
		"no such host",
		"i/o timeout",
		"context deadline exceeded",
		"api 请求失败",
	}
	for _, needle := range needles {
		if strings.Contains(msg, needle) {
			return true
		}
	}
	return false
}
