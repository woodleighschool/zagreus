package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type commandOutput struct {
	activityEnabled     bool
	mu                  sync.Mutex
	out                 io.Writer
	style               textStyle
	logger              *slog.Logger
	progress            *terminalProgress
	interactive, daemon bool
	level               string
	threshold           slog.LevelVar
	verbose             bool
}

func (o *commandOutput) start(cmd *cobra.Command) error {
	o.out = cmd.ErrOrStderr()
	o.style = newTextStyle(o.out)
	o.daemon = cmd.Name() == "run"
	o.threshold.Set(slog.LevelInfo)
	if o.daemon {
		level, err := parseLogLevel(o.level)
		if err != nil {
			return err
		}
		o.threshold.Set(level)
	}
	jsonOutput, _ := cmd.Flags().GetBool("json")
	noProgress, _ := cmd.Flags().GetBool("no-progress")
	o.activityEnabled = !o.daemon && !jsonOutput && !noProgress && cmd.Name() != "schema"
	o.interactive = o.activityEnabled && terminalOutput(o.out) && os.Getenv("CI") == ""
	o.logger = slog.New(&activityHandler{Handler: slog.NewJSONHandler(o.out, &slog.HandlerOptions{Level: &o.threshold}), output: o})
	cmd.SetOut(reportWriter{Writer: cmd.OutOrStdout(), output: o})
	return nil
}

type reportWriter struct {
	io.Writer

	output *commandOutput
}

func (w reportWriter) Write(data []byte) (int, error) {
	w.output.stop()
	return w.Writer.Write(data)
}

func (o *commandOutput) stop() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.interactive = false
	o.activityEnabled = false
	if o.progress != nil {
		o.progress.stop()
		o.progress = nil
	}
}

func (o *commandOutput) finish(cmd *cobra.Command, err error) {
	interrupted := cmd.Context() != nil && errors.Is(context.Cause(cmd.Context()), errInterrupted)
	terminated := cmd.Context() != nil && errors.Is(context.Cause(cmd.Context()), errTerminated)
	if interrupted {
		err = errInterrupted
	}
	o.stop()
	if cmd.Name() != "run" && err != nil {
		message := err.Error()
		style := newTextStyle(cmd.ErrOrStderr())
		label := style.paint("✗", color.FgHiRed)
		if interrupted {
			message, label = "Interrupted.", style.paint("-", color.FgHiYellow)
		}
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), label+" "+humanText(message))
	}
	if o.logger == nil {
		if err == nil {
			return
		}
		o.logger = slog.New(slog.NewJSONHandler(cmd.ErrOrStderr(), nil))
	}
	if err != nil && !interrupted && !terminated {
		o.logger.Error("Command failed", "error", err)
	} else {
		o.logger.Info("Service stopped")
	}
}

type activityHandler struct {
	slog.Handler

	output *commandOutput
	attrs  []slog.Attr
}

func (h *activityHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &activityHandler{Handler: h.Handler.WithAttrs(attrs), output: h.output, attrs: append(append([]slog.Attr(nil), h.attrs...), attrs...)}
}

func (h *activityHandler) WithGroup(name string) slog.Handler {
	return &activityHandler{Handler: h.Handler.WithGroup(name), output: h.output, attrs: h.attrs}
}

func (h *activityHandler) Handle(ctx context.Context, record slog.Record) error {
	a := readActivity(record, h.attrs)
	o := h.output
	if o.daemon {
		if a.verbose && !o.verbose {
			return nil
		}
		if a.stage {
			record.Level = slog.LevelDebug
		}
		if (a.progress || a.status) && !o.verbose {
			return nil
		}
		if !h.Enabled(ctx, record.Level) {
			return nil
		}
		record = plainLog(record, a)
		return h.Handler.Handle(ctx, record)
	}
	o.mu.Lock()
	defer o.mu.Unlock()
	if a.stage || a.progress || a.status {
		if !o.activityEnabled && !o.interactive {
			return nil
		}
		if o.progress == nil {
			o.progress = newTerminalProgress(o.out)
			if !o.interactive {
				o.progress.startPlain()
			}
		}
		o.progress.update(a)
	}
	if record.Level < slog.LevelWarn {
		return nil
	}
	var attrs []string
	read := func(attr slog.Attr) bool {
		attrs = append(attrs, attr.Key+"="+attr.Value.String())
		return true
	}
	for _, attr := range h.attrs {
		read(attr)
	}
	record.Attrs(read)
	message := record.Message
	if len(attrs) > 0 {
		message += "; " + strings.Join(attrs, "; ")
	}
	line := o.style.paint("!", color.FgHiYellow) + " " + humanText(message) + "\n"
	if o.progress != nil {
		return o.progress.write(o.out, line)
	}
	_, err := io.WriteString(o.out, line)
	return err
}
