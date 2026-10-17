package main

import (
	"github.com/pahlevikun/applydiff/internal/runtime"
	"github.com/spf13/cobra"
)

var globalConfirm bool

func requireApprovalGate(cmd *cobra.Command) error {
	return runtime.RequireInteractiveApproval(globalConfirm)
}

func actor() runtime.Actor {
	return runtime.CurrentActor()
}
