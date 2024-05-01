package main

import (
	"github.com/PCManiac/compile_vars"
	"github.com/mrFokin/jrpc"
)

func (h *handler) ping(c jrpc.Context) error {
	return c.Result("pong")
}

func (h *handler) versionCheck(c jrpc.Context) error {
	return c.Result(map[string]interface{}{
		"version":    compile_vars.GetVersion(),
		"build_time": compile_vars.GetBuildTime(),
	})
}
