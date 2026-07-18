package usecases_physical_mysql

import (
	"context"
	"fmt"
	"os/exec"
)

type OSCommandRunner struct{}

func (OSCommandRunner) Run(ctx context.Context, command CommandExecution) error {
	cmd := exec.CommandContext(ctx, command.Executable, command.Args...)
	cmd.Stdin = command.Stdin
	cmd.Stdout = command.Stdout
	cmd.Stderr = command.Stderr

	return cmd.Run()
}

func (OSCommandRunner) RunPipeline(
	ctx context.Context,
	producer CommandExecution,
	consumer CommandExecution,
) error {
	producerCmd := exec.CommandContext(ctx, producer.Executable, producer.Args...)
	producerCmd.Stdin = producer.Stdin
	producerCmd.Stderr = producer.Stderr

	consumerCmd := exec.CommandContext(ctx, consumer.Executable, consumer.Args...)
	consumerCmd.Stdout = consumer.Stdout
	consumerCmd.Stderr = consumer.Stderr

	producerStdout, err := producerCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("open producer stdout pipe: %w", err)
	}
	consumerCmd.Stdin = producerStdout

	if err := consumerCmd.Start(); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}
	if err := producerCmd.Start(); err != nil {
		_ = consumerCmd.Process.Kill()
		_ = consumerCmd.Wait()
		return fmt.Errorf("start producer: %w", err)
	}

	producerErr := producerCmd.Wait()
	consumerErr := consumerCmd.Wait()
	if producerErr != nil {
		return fmt.Errorf("producer failed: %w", producerErr)
	}
	if consumerErr != nil {
		return fmt.Errorf("consumer failed: %w", consumerErr)
	}

	return nil
}
